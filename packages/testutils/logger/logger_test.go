package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	require.NotPanics(t, func() {
		GetDefaultLogger()("test")
	})
}

func TestCustomLogger(t *testing.T) {
	called := false
	fn := func(string, ...interface{}) {
		called = true
	}
	SetDefaultLogger(fn)
	GetDefaultLogger()("test")
	require.True(t, called)
}
