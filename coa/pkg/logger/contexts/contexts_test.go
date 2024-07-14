package contexts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewActivityLogContext(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	assert.Equal(t, "resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "operationName", ctx.operationName)
	assert.Equal(t, "cloudLocation", ctx.cloudLocation)
	assert.Equal(t, "category", ctx.category)
	assert.Equal(t, "correlationId", ctx.correlationId)
	assert.NotNil(t, ctx.properties)
	assert.Equal(t, "callerId", ctx.properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", ctx.properties[Activity_Props_ResourceK8SId])
}

func TestActivityLogContext_ToMap(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	m := ctx.ToMap()
	assert.NotNil(t, m)
	assert.Equal(t, "resourceCloudId", m[Activity_ResourceCloudId])
	assert.Equal(t, "operationName", m[Activity_OperationName])
	assert.Equal(t, "cloudLocation", m[Activity_Location])
	assert.Equal(t, "category", m[Activity_Category])
	assert.Equal(t, "correlationId", m[Activity_CorrelationId])
	assert.NotNil(t, m[Activity_Properties])
	properties := m[Activity_Properties].(map[string]interface{})
	assert.Equal(t, "callerId", properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", properties[Activity_Props_ResourceK8SId])
}

func TestActivityLogContext_FromMap(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	m := map[string]interface{}{
		Activity_ResourceCloudId: "newResourceCloudId",
		Activity_OperationName:   "newOperationName",
		Activity_Location:        "newCloudLocation",
		Activity_Category:        "newCategory",
		Activity_CorrelationId:   "newCorrelationId",
		Activity_Properties: map[string]interface{}{
			Activity_Props_CallerId:      "newCallerId",
			Activity_Props_ResourceK8SId: "newResourceK8SId",
		},
	}
	ctx.FromMap(m)
	assert.Equal(t, "newResourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "newOperationName", ctx.operationName)
	assert.Equal(t, "newCloudLocation", ctx.cloudLocation)
	assert.Equal(t, "newCategory", ctx.category)
	assert.Equal(t, "newCorrelationId", ctx.correlationId)
	assert.NotNil(t, ctx.properties)
	assert.Equal(t, "newCallerId", ctx.properties[Activity_Props_CallerId])
	assert.Equal(t, "newResourceK8SId", ctx.properties[Activity_Props_ResourceK8SId])
}

func TestActivityLogContext_FromMapMissingFields(t *testing.T) {
	ctx := NewActivityLogContext("a_resourceCloudId", "a_cloudLocation", "a_operationName", "a_category", "a_correlationId", "a_callerId", "a_resourceK8SId")
	assert.NotNil(t, ctx)
	m := map[string]interface{}{
		Activity_ResourceCloudId: "resourceCloudId",
	}
	ctx.FromMap(m)
	assert.Equal(t, "resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "a_operationName", ctx.operationName)
	assert.Equal(t, "a_cloudLocation", ctx.cloudLocation)
	assert.Equal(t, "a_category", ctx.category)
	assert.Equal(t, "a_correlationId", ctx.correlationId)
	assert.NotNil(t, ctx.properties)
	assert.Equal(t, "a_callerId", ctx.properties[Activity_Props_CallerId])
	assert.Equal(t, "a_resourceK8SId", ctx.properties[Activity_Props_ResourceK8SId])

	m = map[string]interface{}{
		Activity_Properties: map[string]interface{}{
			Activity_Props_CallerId: "callerId",
		},
	}
	ctx.FromMap(m)
	assert.Equal(t, "resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "a_operationName", ctx.operationName)
	assert.Equal(t, "a_cloudLocation", ctx.cloudLocation)
	assert.Equal(t, "a_category", ctx.category)
	assert.Equal(t, "a_correlationId", ctx.correlationId)
	assert.NotNil(t, ctx.properties)
	assert.Equal(t, "callerId", ctx.properties[Activity_Props_CallerId])
	assert.Equal(t, "a_resourceK8SId", ctx.properties[Activity_Props_ResourceK8SId])

	m = map[string]interface{}{
		Activity_Properties: map[string]interface{}{
			Activity_Props_CallerId:      "callerId",
			Activity_Props_ResourceK8SId: nil,
		},
	}
	ctx.FromMap(m)
	assert.Equal(t, "resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "a_operationName", ctx.operationName)
	assert.Equal(t, "a_cloudLocation", ctx.cloudLocation)
	assert.Equal(t, "a_category", ctx.category)
	assert.Equal(t, "a_correlationId", ctx.correlationId)
	assert.NotNil(t, ctx.properties)
	assert.Equal(t, "callerId", ctx.properties[Activity_Props_CallerId])
	assert.Equal(t, nil, ctx.properties[Activity_Props_ResourceK8SId])
}

func TestActivityLogContext_Deadline(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	deadline, ok := ctx.Deadline()
	assert.False(t, ok)
	assert.Equal(t, deadline, deadline)
}

func TestActivityLogContext_Done(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	done := ctx.Done()
	assert.Nil(t, done)
}

func TestActivityLogContext_Err(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	err := ctx.Err()
	assert.Nil(t, err)
}

func TestActivityLogContext_Value(t *testing.T) {
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	assert.NotNil(t, ctx)
	assert.Equal(t, "resourceCloudId", ctx.Value(Activity_ResourceCloudId))
	assert.Equal(t, "operationName", ctx.Value(Activity_OperationName))
	assert.Equal(t, "cloudLocation", ctx.Value(Activity_Location))
	assert.Equal(t, "category", ctx.Value(Activity_Category))
	assert.Equal(t, "correlationId", ctx.Value(Activity_CorrelationId))
	assert.NotNil(t, ctx.Value(Activity_Properties))
	properties := ctx.Value(Activity_Properties).(map[string]interface{})
	assert.Equal(t, "callerId", properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", properties[Activity_Props_ResourceK8SId])
}

func TestNewDiagnosticLogContext(t *testing.T) {
	ctx := NewDiagnosticLogContext("correlationId", "resourceCloudId", "traceId", "spanId")
	assert.NotNil(t, ctx)
	assert.Equal(t, "correlationId", ctx.correlationId)
	assert.Equal(t, "resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "traceId", ctx.traceContext.traceId)
	assert.Equal(t, "spanId", ctx.traceContext.spanId)
}

func TestDiagnosticsLogContext_ToMap(t *testing.T) {
	ctx := NewDiagnosticLogContext("correlationId", "resourceCloudId", "traceId", "spanId")
	assert.NotNil(t, ctx)
	m := ctx.ToMap()
	assert.NotNil(t, m)
	assert.Equal(t, "correlationId", m[Diagnostics_CorrelationId])
	assert.Equal(t, "resourceCloudId", m[Diagnostics_ResourceCloudId])
	assert.NotNil(t, m[Diagnostics_TraceContext])
	traceContext := m[Diagnostics_TraceContext].(map[string]interface{})
	assert.Equal(t, "traceId", traceContext[Diagnostics_TraceContext_TraceId])
	assert.Equal(t, "spanId", traceContext[Diagnostics_TraceContext_SpanId])
}

func TestDiagnosticsLogContext_FromMap(t *testing.T) {
	ctx := NewDiagnosticLogContext("a_correlationId", "a_resourceCloudId", "a_traceId", "a_spanId")
	assert.NotNil(t, ctx)
	m := map[string]interface{}{
		Diagnostics_CorrelationId:   "correlationId",
		Diagnostics_ResourceCloudId: "resourceCloudId",
		Diagnostics_TraceContext: map[string]interface{}{
			Diagnostics_TraceContext_TraceId: "traceId",
			Diagnostics_TraceContext_SpanId:  "spanId",
		},
	}
	ctx.FromMap(m)
	assert.Equal(t, "correlationId", ctx.correlationId)
	assert.Equal(t, "resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "traceId", ctx.traceContext.traceId)
	assert.Equal(t, "spanId", ctx.traceContext.spanId)
}

func TestDiagnosticsLogContext_FromMapMissingFields(t *testing.T) {
	ctx := NewDiagnosticLogContext("a_correlationId", "a_resourceCloudId", "a_traceId", "a_spanId")
	assert.NotNil(t, ctx)
	m := map[string]interface{}{
		Diagnostics_CorrelationId: "correlationId",
	}
	ctx.FromMap(m)
	assert.Equal(t, "correlationId", ctx.correlationId)
	assert.Equal(t, "a_resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "a_traceId", ctx.traceContext.traceId)
	assert.Equal(t, "a_spanId", ctx.traceContext.spanId)

	m = map[string]interface{}{
		Diagnostics_TraceContext: map[string]interface{}{
			Diagnostics_TraceContext_TraceId: "traceId",
		},
	}
	ctx.FromMap(m)
	assert.Equal(t, "correlationId", ctx.correlationId)
	assert.Equal(t, "a_resourceCloudId", ctx.resourceCloudId)
	assert.Equal(t, "traceId", ctx.traceContext.traceId)
	assert.Equal(t, "a_spanId", ctx.traceContext.spanId)
}

func TestDiagnosticsLogContext_Deadline(t *testing.T) {
	ctx := NewDiagnosticLogContext("traceId", "spanId", "correlationId", "requestId")
	assert.NotNil(t, ctx)
	deadline, ok := ctx.Deadline()
	assert.False(t, ok)
	assert.Equal(t, deadline, deadline)
}

func TestDiagnosticsLogContext_Done(t *testing.T) {
	ctx := NewDiagnosticLogContext("traceId", "spanId", "correlationId", "requestId")
	assert.NotNil(t, ctx)
	done := ctx.Done()
	assert.Nil(t, done)
}

func TestDiagnosticsLogContext_Err(t *testing.T) {
	ctx := NewDiagnosticLogContext("traceId", "spanId", "correlationId", "requestId")
	assert.NotNil(t, ctx)
	err := ctx.Err()
	assert.Nil(t, err)
}

func TestDiagnosticsLogContext_Value(t *testing.T) {
	ctx := NewDiagnosticLogContext("correlationId", "resourceCloudId", "traceId", "spanId")
	assert.NotNil(t, ctx)
	assert.Equal(t, "correlationId", ctx.Value(Diagnostics_CorrelationId))
	assert.Equal(t, "resourceCloudId", ctx.Value(Diagnostics_ResourceCloudId))
	assert.NotNil(t, ctx.Value(Diagnostics_TraceContext))
	traceContext := ctx.Value(Diagnostics_TraceContext).(TraceContext)
	assert.Equal(t, "traceId", traceContext.traceId)
	assert.Equal(t, "spanId", traceContext.spanId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_BackgroundCtx(t *testing.T) {
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("correlationId", "resourceId", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "resourceId", diagCtx.resourceCloudId)
	assert.Equal(t, "correlationId", diagCtx.correlationId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_NilCtx(t *testing.T) {
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("correlationId", "resourceId", nil)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "resourceId", diagCtx.resourceCloudId)
	assert.Equal(t, "correlationId", diagCtx.correlationId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_NilCorrelationId(t *testing.T) {
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("", "resourceId", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "", diagCtx.correlationId)
	assert.Equal(t, "resourceId", diagCtx.resourceCloudId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_NilResourceId(t *testing.T) {
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("correlationId", "", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "", diagCtx.resourceCloudId)
	assert.Equal(t, "correlationId", diagCtx.correlationId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_NilCorrelationIdAndResourceId(t *testing.T) {
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("", "", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "", diagCtx.resourceCloudId)
	assert.Equal(t, "", diagCtx.correlationId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_ParentCtx(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, NewDiagnosticLogContext("a_correlationId", "a_resourceId", "a_traceId", "a_spanId"))
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("correlationId", "resourceId", parent)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "a_traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "a_spanId", diagCtx.traceContext.spanId)
	assert.Equal(t, "correlationId", diagCtx.correlationId)
	assert.Equal(t, "resourceId", diagCtx.resourceCloudId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_ParentCtxWithInvalidDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, "value")
	assert.Equal(t, "value", parent.Value(DiagnosticLogContextKey))
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("correlationId", "resourceId", parent)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "correlationId", diagCtx.correlationId)
	assert.Equal(t, "resourceId", diagCtx.resourceCloudId)
}

func TestPopulateResourceIdAndCorrelationIdToDiagnosticLogContext_ParentCtxWithOtherValues(t *testing.T) {
	parent := context.WithValue(context.Background(), "key", "value")
	ctx := PopulateResourceIdAndCorrelationIdToDiagnosticLogContext("correlationId", "resourceId", parent)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "correlationId", diagCtx.correlationId)
	assert.Equal(t, "resourceId", diagCtx.resourceCloudId)
	assert.Equal(t, "value", ctx.Value("key"))
}

func TestClearResourceIdAndCorrelationIdFromDiagnosticLogContext_NilParent(t *testing.T) {
	ClearResourceIdAndCorrelationIdFromDiagnosticLogContext(nil)
}

func TestClearResourceIdAndCorrelationIdFromDiagnosticLogContext_ParentCtx(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, NewDiagnosticLogContext("a_correlationId", "a_resourceId", "a_traceId", "a_spanId"))
	ClearResourceIdAndCorrelationIdFromDiagnosticLogContext(&parent)
	diagCtx, ok := parent.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "a_traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "a_spanId", diagCtx.traceContext.spanId)
	assert.Equal(t, "", diagCtx.correlationId)
	assert.Equal(t, "", diagCtx.resourceCloudId)
}

func TestClearResourceIdAndCorrelationIdFromDiagnosticLogContext_ParentCtxWithoutDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), "key", "value")
	ClearResourceIdAndCorrelationIdFromDiagnosticLogContext(&parent)
	assert.Equal(t, "value", parent.Value("key"))
}

func TestClearResourceIdAndCorrelationIdFromDiagnosticLogContext_ParentCtxWithInvalidDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, "value")
	ClearResourceIdAndCorrelationIdFromDiagnosticLogContext(&parent)
	assert.Equal(t, "value", parent.Value(DiagnosticLogContextKey))
}

func TTestClearResourceIdAndCorrelationIdFromDiagnosticLogContext_ParentCtxWithNilDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, nil)
	ClearResourceIdAndCorrelationIdFromDiagnosticLogContext(&parent)
	assert.Nil(t, parent.Value(DiagnosticLogContextKey))
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_BackgroundCtx(t *testing.T) {
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("traceId", "spanId", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "spanId", diagCtx.traceContext.spanId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_NilCtx(t *testing.T) {
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("traceId", "spanId", nil)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "spanId", diagCtx.traceContext.spanId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_NilTraceId(t *testing.T) {
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("", "spanId", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "", diagCtx.traceContext.traceId)
	assert.Equal(t, "spanId", diagCtx.traceContext.spanId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_NilSpanId(t *testing.T) {
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("traceId", "", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "", diagCtx.traceContext.spanId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_NilTraceAndSpanId(t *testing.T) {
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("", "", context.Background())
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "", diagCtx.traceContext.traceId)
	assert.Equal(t, "", diagCtx.traceContext.spanId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_ParentCtx(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, NewDiagnosticLogContext("a_correlationId", "a_resourceId", "a_traceId", "a_spanId"))
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("traceId", "spanId", parent)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "spanId", diagCtx.traceContext.spanId)
	assert.Equal(t, "a_correlationId", diagCtx.correlationId)
	assert.Equal(t, "a_resourceId", diagCtx.resourceCloudId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_ParentCtxWithInvalidDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, "value")
	assert.Equal(t, "value", parent.Value(DiagnosticLogContextKey))
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("traceId", "spanId", parent)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "spanId", diagCtx.traceContext.spanId)
}

func TestPopulateTraceAndSpanToDiagnosticLogContext_ParentCtxWithOtherValues(t *testing.T) {
	parent := context.WithValue(context.Background(), "key", "value")
	ctx := PopulateTraceAndSpanToDiagnosticLogContext("traceId", "spanId", parent)
	assert.NotNil(t, ctx)
	diagCtx, ok := ctx.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "traceId", diagCtx.traceContext.traceId)
	assert.Equal(t, "spanId", diagCtx.traceContext.spanId)
	assert.Equal(t, "value", ctx.Value("key"))
}

func TestClearTraceAndSpanFromDiagnosticLogContext_NilParent(t *testing.T) {
	ClearTraceAndSpanFromDiagnosticLogContext(nil)
}

func TestClearTraceAndSpanFromDiagnosticLogContext_ParentCtx(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, NewDiagnosticLogContext("a_correlationId", "a_resourceId", "a_traceId", "a_spanId"))
	ClearTraceAndSpanFromDiagnosticLogContext(&parent)
	diagCtx, ok := parent.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
	assert.True(t, ok)
	assert.NotNil(t, diagCtx)
	assert.Equal(t, "", diagCtx.traceContext.traceId)
	assert.Equal(t, "", diagCtx.traceContext.spanId)
	assert.Equal(t, "a_correlationId", diagCtx.correlationId)
	assert.Equal(t, "a_resourceId", diagCtx.resourceCloudId)
}

func TestClearTraceAndSpanFromDiagnosticLogContext_ParentCtxWithoutDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), "key", "value")
	ClearTraceAndSpanFromDiagnosticLogContext(&parent)
	assert.Equal(t, "value", parent.Value("key"))
}

func TestClearTraceAndSpanFromDiagnosticLogContext_ParentCtxWithInvalidDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, "value")
	ClearTraceAndSpanFromDiagnosticLogContext(&parent)
	assert.Equal(t, "value", parent.Value(DiagnosticLogContextKey))
}

func TestClearTraceAndSpanFromDiagnosticLogContext_ParentCtxWithNilDiagnosticLogContext(t *testing.T) {
	parent := context.WithValue(context.Background(), DiagnosticLogContextKey, nil)
	ClearTraceAndSpanFromDiagnosticLogContext(&parent)
	assert.Nil(t, parent.Value(DiagnosticLogContextKey))
}
