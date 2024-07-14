package contexts

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestActivityLogContextDecorator_Decorate(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithContext(ctx)
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(ActivityLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*ActivityLogContext)

	assert.Equal(t, "resourceCloudId", innerCtx.resourceCloudId)
	assert.Equal(t, "cloudLocation", innerCtx.cloudLocation)
	assert.Equal(t, "operationName", innerCtx.operationName)
	assert.Equal(t, "category", innerCtx.category)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "callerId", innerCtx.properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", innerCtx.properties[Activity_Props_ResourceK8SId])
}

func TestDiagnosticLogContextDecorator_Decorate(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	entry = entry.WithContext(ctx)
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(DiagnosticLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*DiagnosticLogContext)

	assert.Equal(t, "traceId", innerCtx.traceContext.traceId)
	assert.Equal(t, "spanId", innerCtx.traceContext.spanId)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "resourceId", innerCtx.resourceCloudId)
}

func TestActivityLogContextDecorator_DecorateWithKey(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithField(string(ActivityLogContextKey), ctx)
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	innerEntry := entry.Data[string(ActivityLogContextKey)]
	assert.NotNil(t, innerEntry)

	innerCtx := innerEntry.(*ActivityLogContext)

	assert.Equal(t, "resourceCloudId", innerCtx.resourceCloudId)
	assert.Equal(t, "cloudLocation", innerCtx.cloudLocation)
	assert.Equal(t, "operationName", innerCtx.operationName)
	assert.Equal(t, "category", innerCtx.category)
	assert.Equal(t, "correlationId", innerCtx.correlationId)
	assert.Equal(t, "callerId", innerCtx.properties[Activity_Props_CallerId])
	assert.Equal(t, "resourceK8SId", innerCtx.properties[Activity_Props_ResourceK8SId])
}

func TestDiagnosticLogContextDecorator_DecorateWithKey(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	entry = entry.WithField(string(DiagnosticLogContextKey), ctx)
	entry = d.Decorate(entry)
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
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
}

func TestDiagnosticLogContextDecorator_DecorateWithNil(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])
}

func TestActivityLogContextDecorator_DecorateWithInvalidContext(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = entry.WithContext(context.TODO())
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
}

func TestDiagnosticLogContextDecorator_DecorateWithInvalidContext(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry = entry.WithContext(context.TODO())
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])
}

func TestActivityLogContextDecorator_DecorateWithInvalidKey(t *testing.T) {
	d := &ActivityLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewActivityLogContext("resourceCloudId", "cloudLocation", "operationName", "category", "correlationId", "callerId", "resourceK8SId")
	entry = entry.WithField("invalid", ctx)
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(ActivityLogContextKey)])
}

func TestDiagnosticLogContextDecorator_DecorateWithInvalidKey(t *testing.T) {
	d := &DiagnosticLogContextDecorator{}
	entry := logrus.NewEntry(logrus.StandardLogger())
	ctx := NewDiagnosticLogContext("traceId", "spanId", "correlationId", "requestId")
	entry = entry.WithField("invalid", ctx)
	entry = d.Decorate(entry)
	assert.NotNil(t, entry)
	assert.Nil(t, entry.Data[string(DiagnosticLogContextKey)])
}
