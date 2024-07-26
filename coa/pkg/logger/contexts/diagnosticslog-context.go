package contexts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const (
	DIAGNOSTICS_HTTP_HEADER_PREFIX   string = "X-Diagnostics-"
	Diagnostics_CorrelationId        string = "correlationId"
	Diagnostics_ResourceCloudId      string = "resourceId"
	Diagnostics_TraceContext         string = "traceContext"
	Diagnostics_TraceContext_TraceId string = "traceId"
	Diagnostics_TraceContext_SpanId  string = "spanId"
)

type TraceContext struct {
	traceId string
	spanId  string
}

// DiagnosticLogContext is a context that holds diagnostic information.
type DiagnosticLogContext struct {
	correlationId   string
	resourceCloudId string
	traceContext    TraceContext
}

func TraceContextEquals(t1 *TraceContext, t2 *TraceContext) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 == nil || t2 == nil {
		return false
	}
	return t1.traceId == t2.traceId && t1.spanId == t2.spanId
}

func (ctx TraceContext) DeepEquals(other TraceContext) bool {
	return TraceContextEquals(&ctx, &other)
}

func DiagnosticLogContextEquals(d1 *DiagnosticLogContext, d2 *DiagnosticLogContext) bool {
	if d1 == nil && d2 == nil {
		return true
	}
	if d1 == nil || d2 == nil {
		return false
	}
	return d1.correlationId == d2.correlationId &&
		d1.resourceCloudId == d2.resourceCloudId &&
		TraceContextEquals(&d1.traceContext, &d2.traceContext)
}

func (ctx DiagnosticLogContext) DeepEquals(other DiagnosticLogContext) bool {
	return DiagnosticLogContextEquals(&ctx, &other)
}

func (ctx *TraceContext) DeepCopy() *TraceContext {
	if ctx == nil {
		return nil
	}
	return &TraceContext{
		traceId: ctx.traceId,
		spanId:  ctx.spanId,
	}
}

func (ctx *DiagnosticLogContext) DeepCopy() *DiagnosticLogContext {
	if ctx == nil {
		return nil
	}
	return &DiagnosticLogContext{
		correlationId:   ctx.correlationId,
		resourceCloudId: ctx.resourceCloudId,
		traceContext:    *ctx.traceContext.DeepCopy(),
	}
}

func NewDiagnosticLogContext(correlationId, resourceCloudId, traceId, spanId string) *DiagnosticLogContext {
	return &DiagnosticLogContext{
		correlationId:   correlationId,
		resourceCloudId: resourceCloudId,
		traceContext: TraceContext{
			traceId: traceId,
			spanId:  spanId,
		},
	}
}

func (ctx *DiagnosticLogContext) ToMap() map[string]interface{} {
	return map[string]interface{}{
		Diagnostics_CorrelationId:   ctx.correlationId,
		Diagnostics_ResourceCloudId: ctx.resourceCloudId,
		Diagnostics_TraceContext: map[string]interface{}{
			Diagnostics_TraceContext_TraceId: ctx.traceContext.traceId,
			Diagnostics_TraceContext_SpanId:  ctx.traceContext.spanId,
		},
	}
}

func (ctx *DiagnosticLogContext) FromMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	if m[Diagnostics_CorrelationId] != nil {
		ctx.correlationId = m[Diagnostics_CorrelationId].(string)
	}
	if m[Diagnostics_ResourceCloudId] != nil {
		ctx.resourceCloudId = m[Diagnostics_ResourceCloudId].(string)
	}
	if m[Diagnostics_TraceContext] != nil {
		traceContext := m[Diagnostics_TraceContext].(map[string]interface{})
		if traceContext[Diagnostics_TraceContext_TraceId] != nil {
			ctx.traceContext.traceId = traceContext[Diagnostics_TraceContext_TraceId].(string)
		}
		if traceContext[Diagnostics_TraceContext_SpanId] != nil {
			ctx.traceContext.spanId = traceContext[Diagnostics_TraceContext_SpanId].(string)
		}
	}
}

func (ctx *DiagnosticLogContext) String() string {
	b, _ := json.Marshal(ctx.ToMap())
	return string(b)
}

// Deadline returns the time when work done on behalf of this context
func (ctx *DiagnosticLogContext) Deadline() (deadline time.Time, ok bool) {
	// No deadline set
	return time.Time{}, false
}

// Done returns a channel that's closed when work done on behalf of this context should be canceled.
func (ctx *DiagnosticLogContext) Done() <-chan struct{} {
	// No cancellation set
	return nil
}

// Err returns an error if this context has been canceled or timed out.
func (a *DiagnosticLogContext) Err() error {
	// No error set
	return nil
}

// Value returns the value associated with this context for key, or nil if no value is associated with key.
func (ctx *DiagnosticLogContext) Value(key interface{}) interface{} {
	switch key {
	case Diagnostics_CorrelationId:
		return ctx.correlationId
	case Diagnostics_ResourceCloudId:
		return ctx.resourceCloudId
	case Diagnostics_TraceContext:
		return ctx.traceContext
	default:
		return nil
	}
}

func (ctx DiagnosticLogContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.ToMap())
}

func (ctx *DiagnosticLogContext) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	ctx.FromMap(m)
	return nil
}

func (ctx *DiagnosticLogContext) SetCorrelationId(correlationId string) {
	ctx.correlationId = correlationId
}

func (ctx *DiagnosticLogContext) SetResourceId(resourceCloudId string) {
	ctx.resourceCloudId = resourceCloudId
}

func (ctx *DiagnosticLogContext) SetTraceId(traceId string) {
	ctx.traceContext.traceId = traceId
}

func (ctx *DiagnosticLogContext) SetSpanId(spanId string) {
	ctx.traceContext.spanId = spanId
}

func (ctx *DiagnosticLogContext) SetTraceContext(traceContext TraceContext) {
	ctx.traceContext = traceContext
}

func (ctx *DiagnosticLogContext) GetCorrelationId() string {
	return ctx.correlationId
}

func (ctx *DiagnosticLogContext) GetResourceId() string {
	return ctx.resourceCloudId
}

func (ctx *DiagnosticLogContext) GetTraceId() string {
	return ctx.traceContext.traceId
}

func (ctx *DiagnosticLogContext) GetSpanId() string {
	return ctx.traceContext.spanId
}

func (ctx *DiagnosticLogContext) GetTraceContext() TraceContext {
	return ctx.traceContext
}

func OverrideDiagnosticLogContextToCurrentContext(newDiagCtx *DiagnosticLogContext, parent context.Context) context.Context {
	if parent == nil {
		return context.WithValue(context.TODO(), DiagnosticLogContextKey, newDiagCtx)
	}
	return context.WithValue(parent, DiagnosticLogContextKey, newDiagCtx)
}

func PatchDiagnosticLogContextToCurrentContext(newDiagCtx *DiagnosticLogContext, parent context.Context) context.Context {
	if parent == nil {
		return context.WithValue(context.TODO(), DiagnosticLogContextKey, newDiagCtx)
	}
	if diagCtx, ok := parent.Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		// merging
		if newDiagCtx.GetCorrelationId() != "" {
			diagCtx.SetCorrelationId(newDiagCtx.GetCorrelationId())
		}
		if newDiagCtx.GetResourceId() != "" {
			diagCtx.SetResourceId(newDiagCtx.GetResourceId())
		}
		if newDiagCtx.GetTraceId() != "" {
			diagCtx.SetTraceId(newDiagCtx.GetTraceId())
		}
		if newDiagCtx.GetSpanId() != "" {
			diagCtx.SetSpanId(newDiagCtx.GetSpanId())
		}
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	} else {
		return context.WithValue(parent, DiagnosticLogContextKey, newDiagCtx)
	}
}

func PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId string, resourceCloudId string, parent context.Context) context.Context {
	if parent == nil {
		diagCtx := NewDiagnosticLogContext(correlationId, resourceCloudId, "", "")
		return context.WithValue(context.TODO(), DiagnosticLogContextKey, diagCtx)
	}
	if diagCtx, ok := parent.Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		diagCtx.SetCorrelationId(correlationId)
		diagCtx.SetResourceId(resourceCloudId)
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	} else {
		diagCtx := NewDiagnosticLogContext(correlationId, resourceCloudId, "", "")
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	}
}

func ClearResourceIdAndCorrelationIdFromDiagnosticLogContext(parent *context.Context) {
	if parent == nil {
		return
	}
	if diagCtx, ok := (*parent).Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		if diagCtx != nil {
			diagCtx.SetCorrelationId("")
			diagCtx.SetResourceId("")
		}
	}
}

func PopulateTraceAndSpanToDiagnosticLogContext(traceId string, spanId string, parent context.Context) context.Context {
	if parent == nil {
		diagCtx := NewDiagnosticLogContext("", "", traceId, spanId)
		return context.WithValue(context.TODO(), DiagnosticLogContextKey, diagCtx)
	}
	if diagCtx, ok := parent.Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		diagCtx.SetTraceId(traceId)
		diagCtx.SetSpanId(spanId)
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	} else {
		diagCtx := NewDiagnosticLogContext("", "", traceId, spanId)
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	}
}

func ClearTraceAndSpanFromDiagnosticLogContext(parent *context.Context) {
	if parent == nil {
		return
	}
	if diagCtx, ok := (*parent).Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		if diagCtx != nil {
			diagCtx.SetTraceId("")
			diagCtx.SetSpanId("")
		}
	}
}

func ConstructHttpHeaderKeyForDiagnosticsLogContext(key string) string {
	return fmt.Sprintf("%s%s", DIAGNOSTICS_HTTP_HEADER_PREFIX, key)
}

func PropagateDiagnosticLogContextToHttpRequestHeader(req *http.Request) {
	if req == nil {
		return
	}
	if diagCtx, ok := req.Context().Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		req.Header.Set(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_CorrelationId), diagCtx.GetCorrelationId())
		req.Header.Set(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_ResourceCloudId), diagCtx.GetResourceId())
		req.Header.Set(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_TraceContext_TraceId), diagCtx.GetTraceId())
		req.Header.Set(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_TraceContext_SpanId), diagCtx.GetSpanId())
	}
}

func ParseDiagnosticLogContextFromHttpRequestHeader(ctx *fasthttp.RequestCtx) *DiagnosticLogContext {
	if ctx == nil {
		return nil
	}

	correlationId := string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_CorrelationId)))
	resourceCloudId := string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_ResourceCloudId)))
	traceId := string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_TraceContext_TraceId)))
	spanId := string(ctx.Request.Header.Peek(ConstructHttpHeaderKeyForDiagnosticsLogContext(Diagnostics_TraceContext_SpanId)))

	diagCtx := NewDiagnosticLogContext(correlationId, resourceCloudId, traceId, spanId)
	return diagCtx
}

func InheritDiagnosticLogContextFromOriginalContext(orignal context.Context, parent context.Context) context.Context {
	if parent == nil {
		return nil
	}

	if orignal == nil {
		return parent
	}

	if diagCtx, ok := orignal.Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	} else {
		return parent
	}
}

func GenerateCorrelationIdToParentContextIfMissing(parent context.Context) context.Context {
	correlationId := uuid.New().String()
	return PatchCorrelationIdToParentContextIfMissing(parent, correlationId)
}

func PatchCorrelationIdToParentContextIfMissing(parent context.Context, correlationId string) context.Context {
	if parent == nil {
		return nil
	}

	if diagCtx, ok := parent.Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
		if diagCtx.GetCorrelationId() == "" {
			diagCtx.SetCorrelationId(correlationId)
		}
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	} else {
		diagCtx := NewDiagnosticLogContext(correlationId, "", "", "")
		return context.WithValue(parent, DiagnosticLogContextKey, diagCtx)
	}
}
