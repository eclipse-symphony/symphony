package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type (
	compoundingError struct {
		errors []error
		msg    string
	}
)

// Error implements error.
func (c *compoundingError) Error() string {
	finalMessage := strings.Builder{}
	if c.msg != "" {
		finalMessage.WriteString(c.msg)
	} else {
		finalMessage.WriteString("Compounding Error")
	}

	if len(c.errors) != 0 {
		finalMessage.WriteString(":\n")
	}
	for _, err := range c.errors {
		finalMessage.WriteString(fmt.Sprintf("- %v\n", err))
	}

	return finalMessage.String()
}

var _ error = &compoundingError{}

func Eventually(ctx context.Context, condition func(ctx context.Context) error, tick time.Duration, msg string, args ...interface{}) error {
	errs := make([]error, 0)
	errorBuilder := func() error {
		return &compoundingError{
			errors: errs,
			msg:    fmt.Sprintf(msg, args...),
		}
	}

	select {
	case <-ctx.Done():
		return errorBuilder()
	default:
	}

	if err := condition(ctx); err == nil {
		return nil
	} else {
		errs = append(errs, err)
	}

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errorBuilder()
		case <-ticker.C:
			err := condition(ctx)
			if err == nil {
				return nil
			}
			errs = append(errs, err)
		}
	}
}
