package hooks

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewContextHook(t *testing.T) {
	hook := NewContextHook()
	assert.NotNil(t, hook)
	assert.NotNil(t, hook.DiagnosticLogContextDecorator)
	assert.NotNil(t, hook.ActivityLogContextDecorator)
}

func TestContextHook_Fire_WithKeys(t *testing.T) {
	hook := NewContextHook()
	entry := logrus.NewEntry(logrus.StandardLogger())
	diagCtx := contexts.NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	actCtx := contexts.NewActivityLogContext("resourceId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithFields(logrus.Fields{
		string(contexts.DiagnosticLogContextKey): diagCtx,
		string(contexts.ActivityLogContextKey):   actCtx,
	})
	err := hook.Fire(entry)
	assert.Nil(t, err)
	assert.NotNil(t, entry)

	innerEntry := entry.Data[string(contexts.ActivityLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerActCtx := innerEntry.(*contexts.ActivityLogContext)

	assert.Equal(t, "resourceId", innerActCtx.GetResourceCloudId())
	assert.Equal(t, "cloudLocation", innerActCtx.GetCloudLocation())
	assert.Equal(t, "operationName", innerActCtx.GetOperationName())
	assert.Equal(t, "category", innerActCtx.GetCategory())
	assert.Equal(t, "correlationId", innerActCtx.GetCorrelationId())
	assert.Equal(t, "callerId", innerActCtx.GetCallerId())
	assert.Equal(t, "resourceK8SId", innerActCtx.GetResourceK8SId())

	innerEntry = entry.Data[string(contexts.DiagnosticLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerDiagCtx := innerEntry.(*contexts.DiagnosticLogContext)

	assert.Equal(t, "correlationId", innerDiagCtx.GetCorrelationId())
	assert.Equal(t, "resourceId", innerDiagCtx.GetResourceId())
	assert.Equal(t, "traceId", innerDiagCtx.GetTraceId())
	assert.Equal(t, "spanId", innerDiagCtx.GetSpanId())

}

func TestContextHook_Fire_WithActivityLogContext(t *testing.T) {
	hook := NewContextHook()
	entry := logrus.NewEntry(logrus.StandardLogger())
	actCtx := contexts.NewActivityLogContext("resourceId", "cloudLocation", "operationId", "category", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithContext(actCtx)
	err := hook.Fire(entry)
	assert.Nil(t, err)
	assert.NotNil(t, entry)

	innerEntry := entry.Data[string(contexts.ActivityLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerActCtx := innerEntry.(*contexts.ActivityLogContext)

	assert.Equal(t, "resourceId", innerActCtx.GetResourceCloudId())
	assert.Equal(t, "cloudLocation", innerActCtx.GetCloudLocation())
	assert.Equal(t, "operationId", innerActCtx.GetOperationName())
	assert.Equal(t, "category", innerActCtx.GetCategory())
	assert.Equal(t, "correlationId", innerActCtx.GetCorrelationId())
	assert.Equal(t, "callerId", innerActCtx.GetCallerId())
	assert.Equal(t, "resourceK8SId", innerActCtx.GetResourceK8SId())
}

func TestContextHook_Fire_WithDiagnosticLogContext(t *testing.T) {
	hook := NewContextHook()
	entry := logrus.NewEntry(logrus.StandardLogger())
	diagCtx := contexts.NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	entry = entry.WithContext(diagCtx)
	err := hook.Fire(entry)
	assert.Nil(t, err)
	assert.NotNil(t, entry)

	innerEntry := entry.Data[string(contexts.DiagnosticLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerDiagCtx := innerEntry.(*contexts.DiagnosticLogContext)

	assert.Equal(t, "correlationId", innerDiagCtx.GetCorrelationId())
	assert.Equal(t, "resourceId", innerDiagCtx.GetResourceId())
	assert.Equal(t, "traceId", innerDiagCtx.GetTraceId())
	assert.Equal(t, "spanId", innerDiagCtx.GetSpanId())
}

func TestContextHook_Fire_WithOtherContext(t *testing.T) {
	hook := NewContextHook()
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = entry.WithContext(context.TODO())
	err := hook.Fire(entry)
	assert.Nil(t, err)
	assert.NotNil(t, entry)

	assert.Nil(t, entry.Data[string(contexts.ActivityLogContextKey)])
	assert.Nil(t, entry.Data[string(contexts.DiagnosticLogContextKey)])
}
