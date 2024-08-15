package contexts

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
)

// LogContextDecorator is a decorator that parse the context and add to the log entry.
type LogContextDecorator interface {
	// AddFields adds fields to the log entry.
	Decorate(entry *logrus.Entry, folding bool) *logrus.Entry
}

type ActivityLogContextDecorator struct {
}

func (d *ActivityLogContextDecorator) Decorate(entry *logrus.Entry, folding bool) *logrus.Entry {
	var ctx *ActivityLogContext
	var ok bool
	if entry.Context != nil {
		ctx, ok = entry.Context.(*ActivityLogContext)
		if !ok {
			ctx, ok = entry.Context.Value(ActivityLogContextKey).(*ActivityLogContext)
		}
		if ok && ctx != nil {
			if folding {
				entry.Data[string(ActivityLogContextKey)] = ctx
			} else {
				entry.Data[string(OTEL_Activity_ResourceCloudId)] = strings.ToUpper(ctx.GetResourceCloudId())
				entry.Data[string(OTEL_Activity_OperationName)] = ctx.GetOperationName()
				entry.Data[string(OTEL_Activity_Location)] = ctx.GetCloudLocation()
				entry.Data[string(OTEL_Activity_CorrelationId)] = ctx.GetCorrelationId()
				entry.Data[string(OTEL_Activity_Props_CallerId)] = ctx.GetCallerId()
				entry.Data[string(OTEL_Activity_Props_ResourceK8SId)] = ctx.GetResourceK8SId()

				props := ctx.GetProperties()
				filterProps := make(map[string]interface{})
				if props != nil {
					for k, v := range props {
						if k != Activity_Props_CallerId && k != Activity_Props_ResourceK8SId {
							filterProps[k] = v
						}
					}
				}

				filterPropsJson, err := json.Marshal(filterProps)
				if err != nil {
					entry.Data[string(OTEL_Activity_Properties)] = filterProps
				} else {
					entry.Data[string(OTEL_Activity_Properties)] = string(filterPropsJson)
				}
			}
		}
	}
	return entry
}

type DiagnosticLogContextDecorator struct {
}

func (d *DiagnosticLogContextDecorator) Decorate(entry *logrus.Entry, folding bool) *logrus.Entry {
	var ctx *DiagnosticLogContext
	var ok bool
	if entry.Context != nil {
		ctx, ok = entry.Context.(*DiagnosticLogContext)
		if !ok {
			ctx, ok = entry.Context.Value(DiagnosticLogContextKey).(*DiagnosticLogContext)
		}
		if ok && ctx != nil {
			if folding {
				entry.Data[string(DiagnosticLogContextKey)] = ctx
			} else {
				entry.Data[string(OTEL_Diagnostics_CorrelationId)] = ctx.GetCorrelationId()
				entry.Data[string(OTEL_Diagnostics_ResourceCloudId)] = strings.ToUpper(ctx.GetResourceId())

				traceCtxJson, err := json.Marshal(ctx.GetTraceContext())
				if err != nil {
					entry.Data[string(OTEL_Diagnostics_TraceContext)] = ctx.GetTraceContext()
				} else {
					entry.Data[string(OTEL_Diagnostics_TraceContext)] = string(traceCtxJson)
				}
			}
		}
	}
	return entry
}
