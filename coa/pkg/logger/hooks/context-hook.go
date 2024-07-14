package hooks

import (
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/sirupsen/logrus"
)

type ContextHook struct {
	DiagnosticLogContextDecorator *contexts.DiagnosticLogContextDecorator
	ActivityLogContextDecorator   *contexts.ActivityLogContextDecorator
}

func NewContextHook() *ContextHook {
	return &ContextHook{
		DiagnosticLogContextDecorator: &contexts.DiagnosticLogContextDecorator{},
		ActivityLogContextDecorator:   &contexts.ActivityLogContextDecorator{},
	}
}

func (hook *ContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if hook.DiagnosticLogContextDecorator != nil {
			hook.DiagnosticLogContextDecorator.Decorate(entry)
		}
		if hook.ActivityLogContextDecorator != nil {
			hook.ActivityLogContextDecorator.Decorate(entry)
		}
	}
	return nil
}

func (hook *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
