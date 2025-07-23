package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

type MqttBinding struct {
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
	ShouldEnd      string        = "false"
	ConcurrentJobs int           = 3
	Interval       time.Duration = 3 * time.Second
)

type Result struct {
	Result string `json:"result"`
}

var check_response = false
var responseReceived = make(chan bool, 10) // Buffered channel to avoid blocking

var myCorrelationIds sync.Map // store correlationId

// Launch the polling agent
func (m *MqttBinding) Launch() error {
	// Start the agent by handling starter requests
	var get_start = true
	var requests []map[string]interface{}

	// Generate correlationId for initial GET request
	initialCorrelationId := uuid.New().String()
	myCorrelationIds.Store(initialCorrelationId, true)

	Parameters := map[string]string{
		"target":        m.Target,
		"namespace":     m.Namespace,
		"getAll":        "true",
		"preindex":      "0",
		"correlationId": initialCorrelationId,
	}
	request := v1alpha2.COARequest{
		Route:      "tasks",
		Method:     "GET",
		Parameters: Parameters,
	}
	data, _ := json.Marshal(request)
	// Change QoS from 0 to 1 for more reliable delivery
	token := m.Client.Publish(m.RequestTopic, 2, false, data)
	token.Wait()
	var wg sync.WaitGroup
	// subscribe the response topic with QoS 1 instead of 0
	token = m.Client.Subscribe(m.ResponseTopic, 2, func(client mqtt.Client, msg mqtt.Message) {
		var coaResponse v1alpha2.COAResponse
		err := utils2.UnmarshalJson(msg.Payload(), &coaResponse)
		if err != nil {
			fmt.Printf("Error unmarshalling response: %s", err.Error())
			return
		}
		// Parse correlationId from different response structures
		var respMap map[string]interface{}
		_ = json.Unmarshal(coaResponse.Body, &respMap)
		fmt.Printf("Received response: %s\n", string(coaResponse.Body))

		var respCorrelationId string
		correlationKey := contexts.ConstructHttpHeaderKeyForActivityLogContext(contexts.Activity_CorrelationId)

		// Try to get correlationId from top level first (for empty queue responses)
		if topLevelId, ok := respMap[correlationKey].(string); ok && topLevelId != "" {
			respCorrelationId = topLevelId
		}

		if respCorrelationId != "" {
			fmt.Printf("Received response with correlationId: %s\n", respCorrelationId)
			if _, ok := myCorrelationIds.Load(respCorrelationId); !ok {
				// not my request, ignore it
				fmt.Printf("Warning: correlationId is not in map")
				return
			}

		} else {
			fmt.Printf("Warning: correlationId not found in response")
			return
		}
		if coaResponse.State == v1alpha2.BadRequest {
			fmt.Printf("Error: %s\n", string(coaResponse.Body))
			return
		}
		var result Result
		err = utils2.UnmarshalJson(coaResponse.Body, &result)
		if result.Result != "" {
			fmt.Print("handle respponse resultA: %s. \n", result.Result)
			if strings.Contains(result.Result, "handle async result successfully") || strings.Contains(result.Result, "get response successfully") {
				select {
				case responseReceived <- true:
					fmt.Println("Response received successfully")
				default:
					fmt.Println("Response channel is full, skipping")
				}
			} else {
				fmt.Printf("Response not as expected: %s\n", result.Result)
				// Non-blocking send to responseReceived channel
				select {
				case responseReceived <- false:
					fmt.Println("Response received with errors")
				default:
					fmt.Println("Response channel is full, skipping")
				}
			}
		} else {
			fmt.Print("handle respponse B: %+v\n", coaResponse.Body)
			if get_start {
				var allRequests model.ProviderPagingRequest
				err := utils2.UnmarshalJson(coaResponse.Body, &allRequests)
				if err != nil {
					fmt.Printf("Error unmarshalling response: %s", err.Error())
					return
				}

				fmt.Println("Request length: ", len(requests))
				requests = append(requests, allRequests.RequestList...)

				if allRequests.LastMessageID == "" {
					get_start = false
					fmt.Println("get_start: %s", get_start)
					handleRequests(requests, &wg, m)
					fmt.Println("Request length: ", len(requests))
				} else {
					fmt.Println("Request length: ", len(requests))
					// Generate correlationId for continuation request
					continueCorrelationId := uuid.New().String()
					myCorrelationIds.Store(continueCorrelationId, true)

					Parameters := map[string]string{
						"target":        m.Target,
						"namespace":     m.Namespace,
						"getAll":        "true",
						"preindex":      allRequests.LastMessageID,
						"correlationId": continueCorrelationId,
					}
					request := v1alpha2.COARequest{
						Route:      "tasks",
						Method:     "GET",
						Parameters: Parameters,
					}
					data, _ := json.Marshal(request)
					token := m.Client.Publish(m.RequestTopic, 1, false, data)
					token.Wait()
				}
			} else {

				var singleRequest map[string]interface{}
				err := utils2.UnmarshalJson(coaResponse.Body, &singleRequest)
				fmt.Println("single request is here: ", singleRequest)
				if err != nil {
					fmt.Printf("Error unmarshalling response body: %s", err.Error())
					return
				}
				if strings.Contains(string(coaResponse.Body), m.Target) {
					if err != nil {
						fmt.Printf("Error unmarshalling response body: %s", err.Error())
						return
					}
					fmt.Printf("Sub topic00: %s. \n", singleRequest)
					// handle request
					requests = []map[string]interface{}{singleRequest}
					handleRequests(requests, &wg, m)
				}

			}
		}

	})
	token.Wait()
	fmt.Println("Subscribe topic done: ", m.ResponseTopic)

	// handle request one by one
	handleRequests(requests, &wg, m)

	// get new request
	for ShouldEnd == "false" {
		fmt.Println("Begin to pollRequests")
		m.pollRequests()
	}
	return nil
}

func handleRequests(requests []map[string]interface{}, wg *sync.WaitGroup, m *MqttBinding) {
	for _, request := range requests {
		wg.Add(1)
		go func(req map[string]interface{}) {
			defer wg.Done()
			retCtx := context.TODO()
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
			fmt.Println("Publishing response to MQTT %s", data)
			// Change QoS from 0 to 1
			token := m.Client.Publish(m.RequestTopic, 1, false, data)
			// Use WaitTimeout instead of Wait to prevent deadlock
			if !token.WaitTimeout(30 * time.Second) {
				fmt.Println("Warning: MQTT response publish timed out after 30 seconds")
			} else if token.Error() != nil {
				fmt.Printf("Error publishing response: %s\n", token.Error())
			} else {
				fmt.Println("Response published successfully")

				select {
				case success := <-responseReceived:
					if success {
						fmt.Println("Response received successfully")
					} else {
						fmt.Println("Response not received successfully")
					}
				case <-time.After(10 * time.Second):
					fmt.Println("Timeout waiting for response.")
				}
			}
		}(request)
	}
	wg.Wait()
}

func (m *MqttBinding) pollRequests() {
	for i := 0; i < ConcurrentJobs; i++ {
		// Generate correlationId for polling request
		pollCorrelationId := uuid.New().String()
		myCorrelationIds.Store(pollCorrelationId, true)

		// Publish request to get jobs
		Parameters := map[string]string{
			"target":        m.Target,
			"namespace":     m.Namespace,
			"correlationId": pollCorrelationId,
		}
		request := v1alpha2.COARequest{
			Route:      "tasks",
			Method:     "GET",
			Parameters: Parameters,
		}
		fmt.Println("Begin to request topic Get task")
		data, _ := json.Marshal(request)
		token := m.Client.Publish(m.RequestTopic, 1, false, data)

		// Use WaitTimeout instead of Wait to prevent deadlock
		if !token.WaitTimeout(30 * time.Second) {
			fmt.Println("Warning: MQTT publish timed out after 30 seconds")
		} else if token.Error() != nil {
			fmt.Printf("Error publishing to topic %s: %v\n", m.RequestTopic, token.Error())
		}

		time.Sleep(Interval)
	}
}

// UpdateTopology sends topology update request and waits for response
func (m *MqttBinding) UpdateTopology(topologyContent []byte) error {
	fmt.Printf("Updating topology via MQTT for target %s in namespace %s\n", m.Target, m.Namespace)
	m.topologyUpdateChan = make(chan bool, 1)
	responseTimeout := time.After(30 * time.Second)
	updateRequest := v1alpha2.COARequest{
		Method: "POST",
		Route:  "updatetopology",
		Parameters: map[string]string{
			"target":    m.Target,
			"__name":    m.Target, // Explicitly set both target and __name
			"namespace": m.Namespace,
			"Component": "default",
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

	// Subscribe to response topic to receive topology update results
	token := m.Client.Subscribe(m.ResponseTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		if m.topologyUpdateChan != nil {
			fmt.Printf("Received topology update response: %s\n", string(msg.Payload()))

			var coaResponse v1alpha2.COAResponse
			if err := utils2.UnmarshalJson(msg.Payload(), &coaResponse); err == nil {
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
			} else {
				fmt.Printf("Unable to parse response: %v\n", err)
				select {
				case m.topologyUpdateChan <- false:
					fmt.Println("Failure result sent to channel")
				default:
					fmt.Println("Channel is full or closed")
				}
			}
		}
	})

	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to response topic: %v", token.Error())
	}
	fmt.Printf("Successfully subscribed to response topic: %s\n", m.ResponseTopic)

	// Send topology update request
	fmt.Printf("Sending topology update request...\n")
	fmt.Printf("Request topic: %s\n", m.RequestTopic)

	token = m.Client.Publish(m.RequestTopic, 1, false, requestJSON)
	if !token.WaitTimeout(30 * time.Second) {
		return fmt.Errorf("topology update request timed out after 30 seconds")
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to send topology update request: %v", token.Error())
	}

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

	// Clear channel reference but don't unsubscribe - Launch method will reuse this subscription
	m.topologyUpdateChan = nil
	return nil
}
