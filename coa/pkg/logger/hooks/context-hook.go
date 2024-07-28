package hooks

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel/log/global"
)

type ContextHook struct {
	DiagnosticLogContextDecorator *contexts.DiagnosticLogContextDecorator
	ActivityLogContextDecorator   *contexts.ActivityLogContextDecorator
	Folding                       bool
	OtelLogrusHook                *otellogrus.Hook
	OtelLogrusHookEnabled         bool
	OtelLogrusHookLock            sync.RWMutex
	OtelLogrusHookName            string
}

type ContextHookOptions struct {
	DiagnosticLogContextEnabled bool
	ActivityLogContextEnabled   bool
	Folding                     bool
	OtelLogrusHookEnabled       bool
	OtelLogrusHookName          string
}

func NewContextHook() *ContextHook {
	return &ContextHook{
		DiagnosticLogContextDecorator: &contexts.DiagnosticLogContextDecorator{},
		ActivityLogContextDecorator:   &contexts.ActivityLogContextDecorator{},
		Folding:                       true,
	}
}

func NewContextHookWithOptions(options ContextHookOptions) *ContextHook {
	hook := ContextHook{}
	if options.DiagnosticLogContextEnabled {
		hook.DiagnosticLogContextDecorator = &contexts.DiagnosticLogContextDecorator{}
	}
	if options.ActivityLogContextEnabled {
		hook.ActivityLogContextDecorator = &contexts.ActivityLogContextDecorator{}
	}
	hook.Folding = options.Folding
	hook.OtelLogrusHookEnabled = options.OtelLogrusHookEnabled
	hook.OtelLogrusHookName = options.OtelLogrusHookName
	return &hook
}

func (hook *ContextHook) InitializeOtelLogrusHook() {
	if hook.GetOtelLogrusHook() == nil && hook.OtelLogrusHookEnabled && global.GetLoggerProvider() != nil {
		hook.OtelLogrusHookLock.Lock()
		defer hook.OtelLogrusHookLock.Unlock()
		// fmt.Println("Initializing OtelLogrusHook")

		hook.OtelLogrusHook = otellogrus.NewHook(hook.OtelLogrusHookName, otellogrus.WithLevels(hook.Levels()), otellogrus.WithLoggerProvider(global.GetLoggerProvider()))
	}
}

func (hook *ContextHook) GetOtelLogrusHook() *otellogrus.Hook {
	hook.OtelLogrusHookLock.RLock()
	defer hook.OtelLogrusHookLock.RUnlock()
	return hook.OtelLogrusHook
}

func (hook *ContextHook) Fire(entry *logrus.Entry) error {
	// preventing panic in Fire
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Recover panic in ContextHook.Fire, failed with: %v\n", err)
			stacktrace := make([]byte, 8192)
			stackSize := runtime.Stack(stacktrace, false)
			fmt.Printf("%s\n", stacktrace[0:stackSize])
		}
	}()

	if entry.Context != nil {
		if hook.DiagnosticLogContextDecorator != nil {
			hook.DiagnosticLogContextDecorator.Decorate(entry, hook.Folding)
		}
		if hook.ActivityLogContextDecorator != nil {
			hook.ActivityLogContextDecorator.Decorate(entry, hook.Folding)
		}
		if hook.OtelLogrusHookEnabled {
			hook.InitializeOtelLogrusHook()
			if hook.GetOtelLogrusHook() != nil {
				// fmt.Println("Firing entry to OtelLogrusHook")
				hook.GetOtelLogrusHook().Fire(entry)
			}
		}
	}
	return nil
}

func (hook *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
