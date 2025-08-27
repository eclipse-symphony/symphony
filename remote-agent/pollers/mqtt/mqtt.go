package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/eclipse-symphony/symphony/remote-agent/agent"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type MqttPoller struct {
	CertProvider       certs.ICertProvider
	Agent              agent.Agent
	Client             mqtt.Client
	RequestTopic       string
	ResponseTopic      string
	Target             string
	Namespace          string
	topologyUpdateChan chan bool
}

var (
	ConcurrentJobs int           = 3
	Interval       time.Duration = 3 * time.Second
)

type Result struct {
	Result string `json:"result"`
}

// Replace unbounded global sync.Map with a bounded RequestIDCache that supports TTL and periodic cleanup
var requestIDCache *RequestIDCache

// RequestIDCache provides a simple thread-safe cache with TTL and max entries
type RequestIDCache struct {
	mu         sync.RWMutex
	items      map[string]time.Time
	maxEntries int
	ttl        time.Duration
}

const (
	defaultRequestIDTTL        = 5 * time.Minute
	defaultCleanupInterval     = 1 * time.Minute
	defaultMaxRequestIDEntries = 10000
)

func NewRequestIDCache(maxEntries int, ttl time.Duration, cleanupInterval time.Duration) *RequestIDCache {
	c := &RequestIDCache{
		items:      make(map[string]time.Time),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
	// start background cleanup
	go c.cleanupLoop(cleanupInterval)
	return c
}

func (c *RequestIDCache) Store(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[id] = time.Now()
	// enforce maxEntries by removing the oldest entry if exceeded
	if c.maxEntries > 0 && len(c.items) > c.maxEntries {
		var oldestKey string
		var oldest time.Time
		first := true
		for k, t := range c.items {
			if first || t.Before(oldest) {
				oldest = t
				oldestKey = k
				first = false
			}
		}
		delete(c.items, oldestKey)
	}
}

func (c *RequestIDCache) Exists(id string) bool {
	c.mu.RLock()
	t, ok := c.items[id]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	// expire if older than TTL
	if time.Since(t) > c.ttl {
		c.mu.Lock()
		delete(c.items, id)
		c.mu.Unlock()
		return false
	}
	return true
}

func (c *RequestIDCache) Delete(id string) {
	c.mu.Lock()
	delete(c.items, id)
	c.mu.Unlock()
}

func (c *RequestIDCache) PurgeExpired() {
	cutoff := time.Now().Add(-c.ttl)
	c.mu.Lock()
	for k, v := range c.items {
		if v.Before(cutoff) {
			delete(c.items, k)
		}
	}
	// enforce max entries by removing oldest entries if still exceeded
	if c.maxEntries > 0 && len(c.items) > c.maxEntries {
		type kv struct {
			k string
			t time.Time
		}
		arr := make([]kv, 0, len(c.items))
		for k, t := range c.items {
			arr = append(arr, kv{k, t})
		}
		sort.Slice(arr, func(i, j int) bool { return arr[i].t.Before(arr[j].t) })
		over := len(arr) - c.maxEntries
		for i := 0; i < over; i++ {
			delete(c.items, arr[i].k)
		}
	}
	c.mu.Unlock()
}

func (c *RequestIDCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.PurgeExpired()
	}
}

func init() {
	requestIDCache = NewRequestIDCache(defaultMaxRequestIDEntries, defaultRequestIDTTL, defaultCleanupInterval)
}

// Subscribe sets up MQTT response topic subscription
func (m *MqttPoller) Subscribe() error {
	fmt.Println("Setting up MQTT subscription for topic: ", m.ResponseTopic)
	m.Client.Subscribe(m.ResponseTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		var coaResponse v1alpha2.COAResponse
		err := utils2.UnmarshalJson(msg.Payload(), &coaResponse)
		if err != nil {
			fmt.Printf("Error unmarshalling response: %s", err.Error())
			return
		}
		// Check for request-id in response metadata
		var respRequestId string
		if coaResponse.Metadata != nil {
			respRequestId = coaResponse.Metadata["request-id"]
		}

		if respRequestId == "" {
			fmt.Printf("Warning: request-id not found in response metadata")
			return
		}

		fmt.Printf("Received response with request-id: %s\n", respRequestId)
		if !requestIDCache.Exists(respRequestId) {
			// not my request, ignore it
			fmt.Printf("Warning: request-id is not in cache")
			return
		}

		// Handle topology update responses
		if m.topologyUpdateChan != nil {
			fmt.Printf("Received topology update response: %s\n", string(msg.Payload()))
			// Check response status
			if coaResponse.State == v1alpha2.OK || coaResponse.State == v1alpha2.Accepted {
				fmt.Printf("Topology update successful (status code: %d)\n", coaResponse.State)
				select {
				case m.topologyUpdateChan <- true:
					fmt.Println("Success result sent to channel")
				default:
					fmt.Println("Channel is full or closed")
				}
			} else {
				fmt.Printf("Topology update failed (status code: %d): %s\n", coaResponse.State, string(coaResponse.Body))
				select {
				case m.topologyUpdateChan <- false:
					fmt.Println("Failure result sent to channel")
				default:
					fmt.Println("Channel is full or closed")
				}
			}
			// Clean up the request-id from cache
			requestIDCache.Delete(respRequestId)
			return
		}
		if coaResponse.State == v1alpha2.BadRequest {
			fmt.Printf("Error: %s\n", string(coaResponse.Body))
			// Clean up the request-id from cache even on error
			requestIDCache.Delete(respRequestId)
			return
		}

		// Handle regular task responses
		m.handleTaskResponse(coaResponse, respRequestId)
	})

	fmt.Println("Subscribe topic done: ", m.ResponseTopic)
	return nil
}

func (m *MqttPoller) handleTaskResponse(coaResponse v1alpha2.COAResponse, requestId string) {
	// Clean up the request-id from cache when function exits
	defer requestIDCache.Delete(requestId)

	// This function handles task responses and continuation requests for paging
	var requests []map[string]interface{}

	// Try to parse as a paging request first
	var allRequests model.ProviderPagingRequest
	err := utils2.UnmarshalJson(coaResponse.Body, &allRequests)
	if err == nil {
		// Successfully parsed as paging request
		fmt.Printf("Received %d requests from paging response\n", len(allRequests.RequestList))
		requests = append(requests, allRequests.RequestList...)

		// Process current batch of requests with correlation ID
		if len(requests) > 0 {
			// Process requests sequentially
			m.handleRequests(requests)

			// Check if there are more pages to fetch after current page is done
			if allRequests.LastMessageID != "" {
				fmt.Printf("Current page completed. Fetching next page with LastMessageID: %s\n", allRequests.LastMessageID)
				// Generate request-id for continuation request
				continueRequestId := uuid.New().String()
				requestIDCache.Store(continueRequestId)

				Parameters := map[string]string{
					"target":    m.Target,
					"namespace": m.Namespace,
					"getAll":    "true",
					"preindex":  allRequests.LastMessageID,
				}
				request := v1alpha2.COARequest{
					Route:      "tasks",
					Method:     "GET",
					Parameters: Parameters,
					Metadata: map[string]string{
						"request-id": continueRequestId,
					},
				}
				data, _ := json.Marshal(request)
				m.Client.Publish(m.RequestTopic, 0, false, data)
			}
		}
		return
	}

	// If parsing as paging request failed, it might be an empty response or error
	fmt.Printf("No valid requests found in response body: %s\n", string(coaResponse.Body))
}

func (m *MqttPoller) handleRequests(requests []map[string]interface{}) {
	// Process requests sequentially, not concurrently
	for _, request := range requests {
		func(req map[string]interface{}) {
			retCtx := context.TODO()
			// Extract correlation ID from individual request, similar to HTTP poller
			correlationId, ok := req[contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)].(string)
			if !ok {
				fmt.Println("error: correlationId not found in request or not a string. Using a mock one.")
				correlationId = "00000000-0000-0000-0000-000000000000"
			}
			retCtx = context.WithValue(retCtx, contexts.Activity_CorrelationId, correlationId)
			fmt.Printf("Using correlation ID from request: %s\n", correlationId)

			body, err := json.Marshal(req)
			if err != nil {
				fmt.Println("error marshalling request:", err)
				return
			}
			ret := m.Agent.Handle(body, retCtx)
			ret.Namespace = m.Namespace

			// Send response back
			result, err := json.Marshal(ret)
			if err != nil {
				fmt.Println("error marshalling response:", err)
			}
			response := v1alpha2.COARequest{
				Route:       "getResult",
				Method:      "POST",
				ContentType: "application/json",
				Body:        result,
			}
			data, err := json.Marshal(response)
			if err != nil {
				fmt.Printf("Error marshalling response: %s\n", err)
				return
			}
			fmt.Printf("Publishing remote agent execute result to MQTT %s\n", data)
			m.Client.Publish(m.RequestTopic, 0, false, data)
		}(request)
		// Current request completed, continue to next request
	}
}

// Launch the polling agent
func (m *MqttPoller) Launch() error {
	// Start the agent by handling starter requests
	// Generate request-id for initial GET request
	initialRequestId := uuid.New().String()
	requestIDCache.Store(initialRequestId)

	Parameters := map[string]string{
		"target":    m.Target,
		"namespace": m.Namespace,
		"getAll":    "true",
		"preindex":  "0",
	}
	request := v1alpha2.COARequest{
		Route:      "tasks",
		Method:     "GET",
		Parameters: Parameters,
		Metadata: map[string]string{
			"request-id": initialRequestId,
		},
	}
	data, _ := json.Marshal(request)
	m.Client.Publish(m.RequestTopic, 0, false, data)

	// Create ticker for polling
	ticker := time.NewTicker(Interval)
	defer ticker.Stop()

	// Main polling loop - run forever
	for {
		<-ticker.C
		fmt.Println("Begin to pollRequests")
		m.pollRequests()
	}
}

func (m *MqttPoller) pollRequests() {
	for i := 0; i < ConcurrentJobs; i++ {
		// Generate request-id for polling request
		pollRequestId := uuid.New().String()
		requestIDCache.Store(pollRequestId)

		// Publish request to get jobs
		Parameters := map[string]string{
			"target":    m.Target,
			"namespace": m.Namespace,
		}
		request := v1alpha2.COARequest{
			Route:      "tasks",
			Method:     "GET",
			Parameters: Parameters,
			Metadata: map[string]string{
				"request-id": pollRequestId,
			},
		}
		fmt.Println("Begin to request topic Get task")
		data, _ := json.Marshal(request)
		m.Client.Publish(m.RequestTopic, 0, false, data)
	}
}

// UpdateTopology sends topology update request and waits for response
func (m *MqttPoller) UpdateTopology(topologyContent []byte) error {
	fmt.Printf("Updating topology via MQTT for target %s in namespace %s\n", m.Target, m.Namespace)
	m.topologyUpdateChan = make(chan bool, 1)
	responseTimeout := time.After(30 * time.Second)

	// Generate request-id for topology update request
	topologyRequestId := uuid.New().String()
	requestIDCache.Store(topologyRequestId)

	updateRequest := v1alpha2.COARequest{
		Method: "POST",
		Route:  "updatetopology",
		Parameters: map[string]string{
			"target":    m.Target,
			"__name":    m.Target, // Explicitly set both target and __name
			"namespace": m.Namespace,
			"Component": "default",
		},
		Metadata: map[string]string{
			"request-id": topologyRequestId,
		},
		ContentType: "application/json",
		Body:        topologyContent,
	}

	requestJSON, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("failed to serialize topology update request: %v", err)
	}

	if !m.Client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	// Send topology update request without re-subscribing
	// The Launch method's subscription will handle the response
	fmt.Printf("Sending topology update request...\n")
	fmt.Printf("Request topic: %s\n", m.RequestTopic)

	m.Client.Publish(m.RequestTopic, 0, false, requestJSON)

	// Wait for response or timeout
	select {
	case success := <-m.topologyUpdateChan:
		if success {
			fmt.Println("Topology update successfully confirmed")
		} else {
			return fmt.Errorf("topology update failed")
		}
	case <-responseTimeout:
		return fmt.Errorf("timed out waiting for topology update response")
	}

	// Clear channel reference
	m.topologyUpdateChan = nil
	return nil
}
