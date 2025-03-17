package contexts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	Activity_HttpHeaderPrefix           string = "X-Activity-"
	Activity_DiagnosticResourceCloudId  string = "resourceId"
	Activity_DiagnosticResourceLocation string = "location"
	Activity_ResourceCloudId            string = "operatingResourceId"
	Activity_OperationName              string = "operationName"
	Activity_ResourceCloudLocation      string = "operatingResourceLocation"
	Activity_CorrelationId              string = "correlationId"
	Activity_Properties                 string = "properties"
	Activity_EdgeLocation               string = "edgeLocation"

	Activity_Props_CallerId      string = "callerId"
	Activity_Props_ResourceK8SId string = "operatingResourceK8SId"

	OTEL_Activity_DiagnosticResourceCloudId  string = "resourceId"
	OTEL_Activity_DiagnosticResourceLocation string = "location"
	OTEL_Activity_ResourceCloudId            string = "operatingResourceId"
	OTEL_Activity_OperationName              string = "operationName"
	OTEL_Activity_ResourceCloudLocation      string = "operatingResourceLocation"
	OTEL_Activity_CorrelationId              string = "correlationId"
	OTEL_Activity_Properties                 string = "properties"
	OTEL_Activity_Props_CallerId             string = "callerId"
	OTEL_Activity_Props_ResourceK8SId        string = "operatingResourceK8SId"
	OTEL_Activity_Props_EdgeLocation         string = "edgeLocation"
)

// ActivityLogContext is a context that holds activity information.
type ActivityLogContext struct {
	diagnosticResourceCloudId       string
	resourceCloudId                 string
	operationName                   string
	diagnosticResourceCloudLocation string
	resourceCloudLocation           string
	correlationId                   string
	edgeLocation                    string
	properties                      map[string]interface{}
}

func ActivityLogContextEquals(a, b *ActivityLogContext) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.diagnosticResourceCloudId != b.diagnosticResourceCloudId {
		return false
	}
	if a.resourceCloudId != b.resourceCloudId {
		return false
	}
	if a.operationName != b.operationName {
		return false
	}
	if a.diagnosticResourceCloudLocation != b.diagnosticResourceCloudLocation {
		return false
	}
	if a.resourceCloudLocation != b.resourceCloudLocation {
		return false
	}
	if a.correlationId != b.correlationId {
		return false
	}
	if a.edgeLocation != b.edgeLocation {
		return false
	}

	if (a.properties == nil && b.properties != nil) || (a.properties != nil && b.properties == nil) {
		return false
	}

	if a.properties != nil && b.properties != nil {
		if len(a.properties) != len(b.properties) {
			return false
		}
		for k := range a.properties {
			if a.properties[k] != b.properties[k] {
				return false
			}
		}
	}

	return true
}

func (ctx ActivityLogContext) DeepEquals(other ActivityLogContext) bool {
	return ActivityLogContextEquals(&ctx, &other)
}

func (ctx *ActivityLogContext) DeepCopy() *ActivityLogContext {
	if ctx == nil {
		return nil
	}

	newCtx := &ActivityLogContext{
		diagnosticResourceCloudId:       ctx.diagnosticResourceCloudId,
		resourceCloudId:                 ctx.resourceCloudId,
		operationName:                   ctx.operationName,
		diagnosticResourceCloudLocation: ctx.diagnosticResourceCloudLocation,
		resourceCloudLocation:           ctx.resourceCloudLocation,
		correlationId:                   ctx.correlationId,
		edgeLocation:                    ctx.edgeLocation,
		properties:                      make(map[string]interface{}),
	}
	for k := range ctx.properties {
		newCtx.properties[k] = ctx.properties[k]
	}
	return newCtx
}

func NewActivityLogContext(diagnosticResourceCloudId, diagnosticResourceCloudLocation, resourceCloudId, resourceCloudLocation, edgeLocation, operationName, correlationId, callerId, resourceK8SId string) *ActivityLogContext {
	return &ActivityLogContext{
		diagnosticResourceCloudId:       diagnosticResourceCloudId,
		diagnosticResourceCloudLocation: diagnosticResourceCloudLocation,
		resourceCloudId:                 resourceCloudId,
		operationName:                   operationName,
		resourceCloudLocation:           resourceCloudLocation,
		edgeLocation:                    edgeLocation,
		correlationId:                   correlationId,
		properties: map[string]interface{}{
			Activity_Props_CallerId:      callerId,
			Activity_Props_ResourceK8SId: resourceK8SId,
		},
	}
}

func (ctx *ActivityLogContext) ToMap() map[string]interface{} {
	return map[string]interface{}{
		Activity_DiagnosticResourceCloudId:  ctx.diagnosticResourceCloudId,
		Activity_DiagnosticResourceLocation: ctx.diagnosticResourceCloudLocation,
		Activity_ResourceCloudId:            ctx.resourceCloudId,
		Activity_ResourceCloudLocation:      ctx.resourceCloudLocation,
		Activity_OperationName:              ctx.operationName,
		Activity_EdgeLocation:               ctx.edgeLocation,
		Activity_CorrelationId:              ctx.correlationId,
		Activity_Properties:                 ctx.properties,
	}
}

func (ctx *ActivityLogContext) FromMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	if m[Activity_DiagnosticResourceCloudId] != nil {
		ctx.diagnosticResourceCloudId = m[Activity_DiagnosticResourceCloudId].(string)
	}

	if m[Activity_DiagnosticResourceLocation] != nil {
		ctx.diagnosticResourceCloudLocation = m[Activity_DiagnosticResourceLocation].(string)
	}
	if m[Activity_ResourceCloudId] != nil {
		ctx.resourceCloudId = m[Activity_ResourceCloudId].(string)
	}
	if m[Activity_ResourceCloudLocation] != nil {
		ctx.resourceCloudLocation = m[Activity_ResourceCloudLocation].(string)
	}
	if m[Activity_OperationName] != nil {
		ctx.operationName = m[Activity_OperationName].(string)
	}
	if m[Activity_EdgeLocation] != nil {
		ctx.edgeLocation = m[Activity_EdgeLocation].(string)
	}
	if m[Activity_CorrelationId] != nil {
		ctx.correlationId = m[Activity_CorrelationId].(string)
	}
	if m[Activity_Properties] != nil {
		if props, ok := m[Activity_Properties].(map[string]interface{}); ok {
			if ctx.properties == nil {
				ctx.properties = make(map[string]interface{})
			}
			for k := range props {
				ctx.properties[k] = props[k]
			}
		}
	}
}

func (ctx ActivityLogContext) String() string {
	b, _ := json.Marshal(ctx.ToMap())
	return string(b)
}

// Deadline returns the time when work done on behalf of this context
func (ctx *ActivityLogContext) Deadline() (deadline time.Time, ok bool) {
	// No deadline set
	return time.Time{}, false
}

// Done returns a channel that's closed when work done on behalf of this context should be canceled.
func (ctx *ActivityLogContext) Done() <-chan struct{} {
	// No cancellation set
	return nil
}

// Err returns an error if this context has been canceled or timed out.
func (a *ActivityLogContext) Err() error {
	// No error set
	return nil
}

// Value returns the value associated with this context for key, or nil if no value is associated with key.
func (ctx *ActivityLogContext) Value(key interface{}) interface{} {
	switch key {
	case Activity_DiagnosticResourceCloudId:
		return ctx.diagnosticResourceCloudId
	case Activity_DiagnosticResourceLocation:
		return ctx.diagnosticResourceCloudLocation
	case Activity_ResourceCloudId:
		return ctx.resourceCloudId
	case Activity_ResourceCloudLocation:
		return ctx.resourceCloudLocation
	case Activity_OperationName:
		return ctx.operationName
	case Activity_EdgeLocation:
		return ctx.edgeLocation
	case Activity_CorrelationId:
		return ctx.correlationId
	case Activity_Properties:
		return ctx.properties
	default:
		return nil
	}
}

func (ctx ActivityLogContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.ToMap())
}

func (ctx *ActivityLogContext) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	ctx.FromMap(m)
	return nil
}

func (ctx *ActivityLogContext) GetCallerId() string {
	if ctx.properties == nil {
		return ""
	}
	if ctx.properties[Activity_Props_CallerId] == nil {
		return ""
	}
	return ctx.properties[Activity_Props_CallerId].(string)
}

func (ctx *ActivityLogContext) GetDiagnosticResourceCloudId() string {
	return ctx.diagnosticResourceCloudId
}

func (ctx *ActivityLogContext) GetResourceCloudId() string {
	return ctx.resourceCloudId
}

func (ctx *ActivityLogContext) GetResourceK8SId() string {
	if ctx.properties == nil {
		return ""
	}
	if ctx.properties[Activity_Props_ResourceK8SId] == nil {
		return ""
	}
	return ctx.properties[Activity_Props_ResourceK8SId].(string)
}

func (ctx *ActivityLogContext) GetCorrelationId() string {
	return ctx.correlationId
}

func (ctx *ActivityLogContext) GetOperationName() string {
	return ctx.operationName
}

func (ctx *ActivityLogContext) GetDiagnosticResourceCloudLocation() string {
	return ctx.diagnosticResourceCloudLocation
}

func (ctx *ActivityLogContext) GetResourceCloudLocation() string {
	return ctx.resourceCloudLocation
}

func (ctx *ActivityLogContext) GetEdgeLocation() string {
	return ctx.edgeLocation
}

func (ctx *ActivityLogContext) SetCallerId(callerId string) {
	if ctx.properties == nil {
		ctx.properties = make(map[string]interface{})
	}
	ctx.properties[Activity_Props_CallerId] = callerId
}

func (ctx *ActivityLogContext) SetResourceCloudId(resourceCloudId string) {
	ctx.resourceCloudId = resourceCloudId
}

func (ctx *ActivityLogContext) SetDiagnosticResourceCloudId(diagnosticResourceCloudId string) {
	ctx.diagnosticResourceCloudId = diagnosticResourceCloudId
}

func (ctx *ActivityLogContext) SetResourceK8SId(resourceK8SId string) {
	if ctx.properties == nil {
		ctx.properties = make(map[string]interface{})
	}
	ctx.properties[Activity_Props_ResourceK8SId] = resourceK8SId
}

func (ctx *ActivityLogContext) SetCorrelationId(correlationId string) {
	ctx.correlationId = correlationId
}

func (ctx *ActivityLogContext) SetOperationName(operationName string) {
	ctx.operationName = operationName
}

func (ctx *ActivityLogContext) SetDiagnosticResourceCloudLocation(cloudLocation string) {
	ctx.diagnosticResourceCloudLocation = cloudLocation
}

func (ctx *ActivityLogContext) SetEdgeLocation(edgeLocation string) {
	ctx.edgeLocation = edgeLocation
}

func (ctx *ActivityLogContext) SetResourceCloudLocation(resourceCloudLocation string) {
	ctx.resourceCloudLocation = resourceCloudLocation
}

func (ctx *ActivityLogContext) SetProperties(properties map[string]interface{}) {
	ctx.properties = properties
}

func (ctx *ActivityLogContext) GetProperties() map[string]interface{} {
	return ctx.properties
}

func (ctx *ActivityLogContext) SetProperty(key string, value interface{}) {
	if ctx.properties == nil {
		ctx.properties = make(map[string]interface{})
	}
	ctx.properties[key] = value
}

func (ctx *ActivityLogContext) GetProperty(key string) interface{} {
	if ctx.properties == nil {
		return nil
	}
	return ctx.properties[key]
}

func OverrideActivityLogContextToCurrentContext(newActCtx *ActivityLogContext, parent context.Context) context.Context {
	if parent == nil {
		return context.WithValue(context.TODO(), ActivityLogContextKey, newActCtx)
	}
	return context.WithValue(parent, ActivityLogContextKey, newActCtx)
}

func PatchActivityLogContextToCurrentContext(newActCtx *ActivityLogContext, parent context.Context) context.Context {
	if parent == nil {
		return context.WithValue(context.TODO(), ActivityLogContextKey, newActCtx)
	}
	if actCtx, ok := parent.Value(ActivityLogContextKey).(*ActivityLogContext); ok {
		// merging
		if newActCtx.diagnosticResourceCloudId != "" {
			actCtx.SetDiagnosticResourceCloudId(actCtx.diagnosticResourceCloudId)
		}
		if newActCtx.resourceCloudId != "" {
			actCtx.SetResourceCloudId(actCtx.resourceCloudId)
		}
		if newActCtx.operationName != "" {
			actCtx.SetOperationName(actCtx.operationName)
		}
		if newActCtx.diagnosticResourceCloudLocation != "" {
			actCtx.SetDiagnosticResourceCloudLocation(actCtx.diagnosticResourceCloudLocation)
		}
		if newActCtx.resourceCloudLocation != "" {
			actCtx.SetResourceCloudLocation(actCtx.resourceCloudLocation)
		}
		if newActCtx.edgeLocation != "" {
			actCtx.SetEdgeLocation(actCtx.edgeLocation)
		}
		if newActCtx.correlationId != "" {
			actCtx.SetCorrelationId(actCtx.correlationId)
		}
		if newActCtx.properties != nil {
			for k := range newActCtx.properties {
				if v := newActCtx.properties[k]; v != nil && v != "" {
					actCtx.SetProperty(k, v)
				}
			}
		}
		return context.WithValue(parent, ActivityLogContextKey, actCtx)
	} else {
		return context.WithValue(parent, ActivityLogContextKey, newActCtx)
	}
}

func ConstructHttpHeaderKeyForActivityLogContext(key string) string {
	return fmt.Sprintf("%s%s", Activity_HttpHeaderPrefix, key)
}

func PropagateActivityLogContextToHttpRequestHeader(req *http.Request) {
	if req == nil {
		return
	}

	if actCtx, ok := req.Context().Value(ActivityLogContextKey).(*ActivityLogContext); ok {
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceCloudId), actCtx.GetDiagnosticResourceCloudId())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceLocation), actCtx.GetDiagnosticResourceCloudLocation())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId), actCtx.GetResourceCloudId())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudLocation), actCtx.GetResourceCloudLocation())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName), actCtx.GetOperationName())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_EdgeLocation), actCtx.GetEdgeLocation())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_CorrelationId), actCtx.GetCorrelationId())

		props := actCtx.GetProperties()
		propsJson, err := json.Marshal(props)
		if err != nil {
			// skip
		} else {
			req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_Properties), string(propsJson))

		}
	}
}

func IsActivityLogContextPropertiesHeader(key string) bool {
	return key == ConstructHttpHeaderKeyForActivityLogContext(Activity_Properties)
}

func ParseActivityLogContextFromHttpRequestHeader(ctx *fasthttp.RequestCtx) *ActivityLogContext {
	if ctx == nil {
		return nil
	}

	actCtx := ActivityLogContext{}
	actCtx.SetDiagnosticResourceCloudId(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceCloudId))))
	actCtx.SetDiagnosticResourceCloudLocation(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceLocation))))
	actCtx.SetResourceCloudId(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId))))
	actCtx.SetOperationName(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName))))
	actCtx.SetResourceCloudLocation(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudLocation))))
	actCtx.SetEdgeLocation(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_EdgeLocation))))
	actCtx.SetCorrelationId(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_CorrelationId))))

	props := make(map[string]interface{})
	propsJson := string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_Properties)))
	if propsJson != "" {
		if err := json.Unmarshal([]byte(propsJson), &props); err != nil {
			// skip
		} else {
			actCtx.SetProperties(props)
		}
	}

	return &actCtx
}

func InheritActivityLogContextFromOriginalContext(original context.Context, parent context.Context) context.Context {
	if parent == nil {
		return nil
	}

	if original == nil {
		return parent
	}

	if actCtx, ok := original.Value(ActivityLogContextKey).(*ActivityLogContext); ok {
		return context.WithValue(parent, ActivityLogContextKey, actCtx)
	} else {
		return parent
	}
}

func PropagateActivityLogContextToMetadata(ctx context.Context, metadata map[string]string) {
	if ctx == nil {
		return
	}

	if actCtx, ok := ctx.Value(ActivityLogContextKey).(*ActivityLogContext); ok {
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceCloudId)] = actCtx.GetDiagnosticResourceCloudId()
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceLocation)] = actCtx.GetDiagnosticResourceCloudLocation()
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId)] = actCtx.GetResourceCloudId()
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudLocation)] = actCtx.GetResourceCloudLocation()
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName)] = actCtx.GetOperationName()
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_EdgeLocation)] = actCtx.GetEdgeLocation()
		metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_CorrelationId)] = actCtx.GetCorrelationId()

		props := actCtx.GetProperties()
		propsJson, err := json.Marshal(props)
		if err != nil {
			// skip
		} else {
			metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_Properties)] = string(propsJson)
		}
	}
}

func ParseActivityLogContextFromMetadata(metadata map[string]string) *ActivityLogContext {
	if metadata == nil {
		return nil
	}

	actCtx := ActivityLogContext{}
	actCtx.SetDiagnosticResourceCloudId(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceCloudId)])
	actCtx.SetDiagnosticResourceCloudLocation(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceLocation)])
	actCtx.SetResourceCloudId(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId)])
	actCtx.SetResourceCloudLocation(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudLocation)])
	actCtx.SetOperationName(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName)])
	actCtx.SetEdgeLocation(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_EdgeLocation)])
	actCtx.SetCorrelationId(metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_CorrelationId)])

	props := make(map[string]interface{})
	propsJson := metadata[ConstructHttpHeaderKeyForActivityLogContext(Activity_Properties)]
	if propsJson != "" {
		if err := json.Unmarshal([]byte(propsJson), &props); err != nil {
			// skip
		} else {
			actCtx.SetProperties(props)
		}
	}

	return &actCtx
}

func ClearActivityLogContextFromMetadata(metadata map[string]string) {
	if metadata == nil {
		return
	}

	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceCloudId))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_DiagnosticResourceLocation))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudLocation))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_EdgeLocation))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_CorrelationId))
	delete(metadata, ConstructHttpHeaderKeyForActivityLogContext(Activity_Properties))
}
