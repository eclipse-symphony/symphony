package contexts

import "github.com/sirupsen/logrus"

// LogContextDecorator is a decorator that parse the context and add to the log entry.
type LogContextDecorator interface {
	// AddFields adds fields to the log entry.
	Decorate(entry *logrus.Entry) *logrus.Entry
}

type ActivityLogContextDecorator struct {
}

func (d *ActivityLogContextDecorator) Decorate(entry *logrus.Entry) *logrus.Entry {
	if entry.Context != nil {
		if ctx, ok := entry.Context.(*ActivityLogContext); ok {
			entry.Data[string(ActivityLogContextKey)] = ctx
		} else if ctx, ok := entry.Context.Value(ActivityLogContextKey).(*ActivityLogContext); ok {
			entry.Data[string(ActivityLogContextKey)] = ctx
		}
	}
	return entry
}

type DiagnosticLogContextDecorator struct {
}

func (d *DiagnosticLogContextDecorator) Decorate(entry *logrus.Entry) *logrus.Entry {
	if entry.Context != nil {
		if ctx, ok := entry.Context.(*DiagnosticLogContext); ok {
			entry.Data[string(DiagnosticLogContextKey)] = ctx
		} else if ctx, ok := entry.Context.Value(DiagnosticLogContextKey).(*DiagnosticLogContext); ok {
			entry.Data[string(DiagnosticLogContextKey)] = ctx
		}
	}
	return entry
}
