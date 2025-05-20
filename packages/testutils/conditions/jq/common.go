package jq

// Equality returns a new JqCondition that checks if the value at the given path resolves to the given value exactly.
func Equality(path string, value interface{}, opts ...Option) *JqCondition {
	return MustNew(path, append(opts, WithValue(value))...)
}
