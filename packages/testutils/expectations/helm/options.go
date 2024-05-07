package helm

import "github.com/eclipse-symphony/symphony/packages/testutils/types"

// WithRemoved specifies whether the release is expected to be present or not.
func WithRemoved(removed bool) Option {
	return func(h *HelmExpectation) {
		h.removed = removed
	}
}

func WithListClientBuilder(builder func() (ListRunner, error)) Option {
	return func(h *HelmExpectation) {
		h.actionBuilder = builder
	}
}

func WithValueListCondition(condition types.Condition) Option {
	return func(h *HelmExpectation) {
		newC := createValueConditionFrom(condition, true)
		addCondition(&h.releaseListCondition, newC)
	}
}
func WithValueCondition(condition types.Condition) Option {
	return func(h *HelmExpectation) {
		newC := createValueConditionFrom(condition, false)
		addCondition(&h.releaseCondition, newC)
	}
}

func WithReleaseCondition(condition types.Condition) Option {
	return func(h *HelmExpectation) {
		addCondition(&h.releaseCondition, condition)
	}
}

func WithReleaseListCondition(condition types.Condition) Option {
	return func(h *HelmExpectation) {
		addCondition(&h.releaseListCondition, condition)
	}
}

func WithDescription(description string) Option {
	return func(h *HelmExpectation) {
		h.description = description
	}
}

func WithLogger(logger func(format string, args ...interface{})) Option {
	return func(h *HelmExpectation) {
		h.l = logger
	}
}
