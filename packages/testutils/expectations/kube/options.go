package kube

import (
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

// WithCondition specifies the conditions that the resource should satisfy.
func WithCondition(condition types.Condition) Option {
	return func(re *KubeExpectation) {
		re.condition = condition
	}
}

// WithListCondition specifies the conditions that the list of matched resources should satisfy.
func WithListCondition(condition types.Condition) Option {
	return func(re *KubeExpectation) {
		re.listCondition = condition
	}
}

// WithLogger specifies the logger to be used.
func WithLogger(logger func(format string, args ...interface{})) Option {
	return func(re *KubeExpectation) {
		re.l = logger
	}
}

// WithTick specifies the tick for the expectation.
func WithTick(tick time.Duration) Option {
	return func(re *KubeExpectation) {
		re.tick = tick
	}
}

func WithDescription(description string) Option {
	return func(h *KubeExpectation) {
		h.description = description
	}
}

func WithDynamicClientBuilder(builder func() (dynamic.Interface, error)) Option {
	return func(h *KubeExpectation) {
		h.dynamicClientBuilder = builder
	}
}

func WithDiscoveryClientBuilder(builder func() (discovery.DiscoveryInterface, error)) Option {
	return func(h *KubeExpectation) {
		h.discoveryClientBuilder = builder
	}
}

func IsAbsent() Option {
	return func(re *KubeExpectation) {
		re.removed = true
	}
}
