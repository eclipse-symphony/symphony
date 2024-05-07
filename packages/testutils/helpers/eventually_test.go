package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEventually_Success(t *testing.T) {
	condition := func(ctx context.Context) error {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := Eventually(ctx, condition, 1*time.Millisecond, "failure message")
	require.NoError(t, err)
}

func TestEventually_SuccessfulOnSubsequentAttempt(t *testing.T) {
	count := 0
	condition := func(ctx context.Context) error {
		defer func() { count++ }()
		if count <= 1 {
			return errors.New("test error")
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := Eventually(ctx, condition, 1*time.Millisecond, "failure message")
	require.NoError(t, err)
	require.Equal(t, 3, count)
}

func TestEventually_Failure(t *testing.T) {
	condition := func(ctx context.Context) error {
		return errors.New("test error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := Eventually(ctx, condition, 1*time.Millisecond, "failure message")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failure message")
	require.Contains(t, err.Error(), "test error")
}

func TestEventually_FailIfContextIsAlreadyDone(t *testing.T) {
	condition := func(ctx context.Context) error {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Eventually(ctx, condition, 1*time.Millisecond, "failure message")
	require.Error(t, err)
}
