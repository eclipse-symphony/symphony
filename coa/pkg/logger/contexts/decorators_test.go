package contexts

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestActivityLogContextDecorator_Decorate(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("diagnosticResourceId", "diagnosticResourceCloudLocation", "resourceCloudId", "resourceCloudLocation", "edgeLocation", "operationName", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithContext(ctx)
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(ActivityLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*ActivityLogContext)

	assert.Equal(t, "diagnosticResourceId", innerCtx.diagnosticResourceCloudId)
	assert.Equal(t, "diagnosticResourceCloudLocation", innerCtx.diagnosticResourceCloudLocation)
	assert.Equal(t, "resourceCloudId", innerCtx.resourceCloudId)
	assert.Equal(t, "resourceCloudLocation", innerCtx.resourceCloudLocation)
	assert.Equal(t, "edgeLocation", innerCtx.edgeLocation)
	assert.Equal(t, "operationName", innerCtx.operationName)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "callerId", innerCtx.properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", innerCtx.properties[Activity_Props_ResourceK8SId])
}

func TestActivityLogContextDecorator_DecorateUnfold(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("diagnosticResourceId", "diagnosticResourceCloudLocation", "resourceCloudId", "resourceCloudLocation", "edgeLocation", "operationName", "correlationId", "callerId", "resourceK8SId")
	ctx.SetProperty("key", "value")
	entry = entry.WithContext(ctx)
	entry = d.Decorate(entry, false)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
	assert.Equal(t, strings.ToUpper("diagnosticResourceId"), entry.Data[string(OTEL_Activity_DiagnosticResourceCloudId)])
	assert.Equal(t, "diagnosticResourceCloudLocation", entry.Data[string(OTEL_Activity_DiagnosticResourceLocation)])
	assert.Equal(t, strings.ToUpper("resourceCloudId"), entry.Data[string(OTEL_Activity_ResourceCloudId)])
	assert.Equal(t, "edgeLocation", entry.Data[string(OTEL_Activity_Props_EdgeLocation)])
	assert.Equal(t, "operationName", entry.Data[string(OTEL_Activity_OperationName)])
	assert.Equal(t, "correlationId", entry.Data[string(OTEL_Activity_CorrelationId)])
	assert.Equal(t, "callerId", entry.Data[string(OTEL_Activity_Props_CallerId)])
	assert.Equal(t, "resourceK8SId", entry.Data[string(OTEL_Activity_Props_ResourceK8SId)])

	propsJson, ok := entry.Data[string(OTEL_Activity_Properties)].(string)
	assert.True(t, ok)
	var props map[string]interface{}
	err := json.Unmarshal([]byte(propsJson), &props)
	assert.Nil(t, err)
	assert.Equal(t, "value", props["key"])
}

func TestDiagnosticLogContextDecorator_Decorate(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	entry = entry.WithContext(ctx)
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(DiagnosticLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*DiagnosticLogContext)

	assert.Equal(t, "traceId", innerCtx.traceContext.traceId)
	assert.Equal(t, "spanId", innerCtx.traceContext.spanId)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "resourceId", innerCtx.resourceCloudId)
}

func TestDiagnosticLogContextDecorator_DecorateUnfold(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	entry = entry.WithContext(ctx)
	entry = d.Decorate(entry, false)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])

	assert.Equal(t, "correlationId", entry.Data[string(OTEL_Diagnostics_CorrelationId)])
	assert.Equal(t, strings.ToUpper("resourceId"), entry.Data[string(OTEL_Diagnostics_ResourceCloudId)])
	traceCtxJson, ok := entry.Data[string(OTEL_Diagnostics_TraceContext)].(string)
	assert.True(t, ok)
	var traceCtx TraceContext
	err := json.Unmarshal([]byte(traceCtxJson), &traceCtx)
	assert.Nil(t, err)
	assert.Equal(t, "traceId", traceCtx.traceId)
	assert.Equal(t, "spanId", traceCtx.spanId)
}

func TestActivityLogContextDecorator_DecorateWithKey(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("diagnosticResourceId", "diagnosticResourceCloudLocation", "resourceCloudId", "resourceCloudLocation", "edgeLocation", "operationName", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithField(string(ActivityLogContextKey), ctx)
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(ActivityLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*ActivityLogContext)

	assert.Equal(t, "diagnosticResourceId", innerCtx.diagnosticResourceCloudId)
	assert.Equal(t, "diagnosticResourceCloudLocation", innerCtx.diagnosticResourceCloudLocation)
	assert.Equal(t, "resourceCloudId", innerCtx.resourceCloudId)
	assert.Equal(t, "resourceCloudLocation", innerCtx.resourceCloudLocation)
	assert.Equal(t, "edgeLocation", innerCtx.edgeLocation)
	assert.Equal(t, "operationName", innerCtx.operationName)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "callerId", innerCtx.properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", innerCtx.properties[Activity_Props_ResourceK8SId])
}

func TestDiagnosticLogContextDecorator_DecorateWithKey(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	entry = entry.WithField(string(DiagnosticLogContextKey), ctx)
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(DiagnosticLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*DiagnosticLogContext)

	assert.Equal(t, "traceId", innerCtx.traceContext.traceId)
	assert.Equal(t, "spanId", innerCtx.traceContext.spanId)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "resourceId", innerCtx.resourceCloudId)
}

func TestActivityLogContextDecorator_DecorateWithNil(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
}

func TestDiagnosticLogContextDecorator_DecorateWithNil(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])
}

func TestActivityLogContextDecorator_DecorateWithInvalidContext(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = entry.WithContext(context.TODO())
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
}

func TestDiagnosticLogContextDecorator_DecorateWithInvalidContext(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = entry.WithContext(context.TODO())
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])
}

func TestActivityLogContextDecorator_DecorateWithInvalidKey(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("diagnosticResourceId", "diagnosticResourceCloudLocation", "resourceCloudId", "resourceCloudLocation", "edgeLocation", "operationName", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithField("invalid", ctx)
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
}

func TestDiagnosticLogContextDecorator_DecorateWithInvalidKey(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("traceId", "spanId", "correlationId", "requestId")
	entry = entry.WithField("invalid", ctx)
	entry = d.Decorate(entry, true)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])
}
