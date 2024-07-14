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
	ACTIVITY_HTTP_HEADER_PREFIX string = "X-Activity-"
	Activity_ResourceCloudId    string = "resourceId"
	Activity_OperationName      string = "operationName"
	Activity_Location           string = "location"
	Activity_Category           string = "category"
	Activity_CorrelationId      string = "correlationId"
	Activity_Properties         string = "properties"

	Activity_Props_CallerId      string = "caller-id"
	Activity_Props_ResourceK8SId string = "resource-k8s-id"
)

// ActivityLogContext is a context that holds activity information.
type ActivityLogContext struct {
	resourceCloudId string
	operationName   string
	cloudLocation   string
	category        string
	correlationId   string
	properties      map[string]interface{}
}

func ActivityLogContextEquals(a, b *ActivityLogContext) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.resourceCloudId != b.resourceCloudId {
		return false
	}
	if a.operationName != b.operationName {
		return false
	}
	if a.cloudLocation != b.cloudLocation {
		return false
	}
	if a.category != b.category {
		return false
	}
	if a.correlationId != b.correlationId {
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
		resourceCloudId: ctx.resourceCloudId,
		operationName:   ctx.operationName,
		cloudLocation:   ctx.cloudLocation,
		category:        ctx.category,
		correlationId:   ctx.correlationId,
		properties:      make(map[string]interface{}),
	}
	for k := range ctx.properties {
		newCtx.properties[k] = ctx.properties[k]
	}
	return newCtx
}

func NewActivityLogContext(resourceCloudId, cloudLocation, operationName, category, correlationId, callerId, resourceK8SId string) *ActivityLogContext {
	return &ActivityLogContext{
		resourceCloudId: resourceCloudId,
		operationName:   operationName,
		cloudLocation:   cloudLocation,
		category:        category,
		correlationId:   correlationId,
		properties: map[string]interface{}{
			Activity_Props_CallerId:      callerId,
			Activity_Props_ResourceK8SId: resourceK8SId,
		},
	}
}

func (ctx *ActivityLogContext) ToMap() map[string]interface{} {
	return map[string]interface{}{
		Activity_ResourceCloudId: ctx.resourceCloudId,
		Activity_OperationName:   ctx.operationName,
		Activity_Location:        ctx.cloudLocation,
		Activity_Category:        ctx.category,
		Activity_CorrelationId:   ctx.correlationId,
		Activity_Properties:      ctx.properties,
	}
}

func (ctx *ActivityLogContext) FromMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	if m[Activity_ResourceCloudId] != nil {
		ctx.resourceCloudId = m[Activity_ResourceCloudId].(string)
	}
	if m[Activity_OperationName] != nil {
		ctx.operationName = m[Activity_OperationName].(string)
	}
	if m[Activity_Location] != nil {
		ctx.cloudLocation = m[Activity_Location].(string)
	}
	if m[Activity_Category] != nil {
		ctx.category = m[Activity_Category].(string)
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

func (ctx *ActivityLogContext) String() string {
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
	case Activity_ResourceCloudId:
		return ctx.resourceCloudId
	case Activity_OperationName:
		return ctx.operationName
	case Activity_Location:
		return ctx.cloudLocation
	case Activity_Category:
		return ctx.category
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

func (ctx *ActivityLogContext) GetCloudLocation() string {
	return ctx.cloudLocation
}

func (ctx *ActivityLogContext) GetCategory() string {
	return ctx.category
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

func (ctx *ActivityLogContext) SetCloudLocation(cloudLocation string) {
	ctx.cloudLocation = cloudLocation
}

func (ctx *ActivityLogContext) SetCategory(category string) {
	ctx.category = category
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
		if newActCtx.resourceCloudId != "" {
			actCtx.SetResourceCloudId(actCtx.resourceCloudId)
		}
		if newActCtx.operationName != "" {
			actCtx.SetOperationName(actCtx.operationName)
		}
		if newActCtx.cloudLocation != "" {
			actCtx.SetCloudLocation(actCtx.cloudLocation)
		}
		if newActCtx.category != "" {
			actCtx.SetCategory(actCtx.category)
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
	return fmt.Sprintf("%s%s", ACTIVITY_HTTP_HEADER_PREFIX, key)
}

func PropagateActivityLogContextToHttpRequestHeader(req *http.Request) {
	if req == nil {
		return
	}

	if actCtx, ok := req.Context().Value(ActivityLogContextKey).(*ActivityLogContext); ok {
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId), actCtx.GetResourceCloudId())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName), actCtx.GetOperationName())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_Location), actCtx.GetCloudLocation())
		req.Header.Set(ConstructHttpHeaderKeyForActivityLogContext(Activity_Category), actCtx.GetCategory())
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
	actCtx.SetResourceCloudId(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_ResourceCloudId))))
	actCtx.SetOperationName(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_OperationName))))
	actCtx.SetCloudLocation(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_Location))))
	actCtx.SetCategory(string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForActivityLogContext(Activity_Category))))
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
