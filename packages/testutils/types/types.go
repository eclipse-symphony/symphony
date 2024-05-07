package types

import (
	"context"

	"github.com/onsi/gomega/types"
)

type (
	id interface {
		Id() string
		Description() string
	}

	// Expectation is a way to define expectations for a resource. It is used to verify whether a resource is in the
	// expected state. It's responsible for retrieving the resource and verifying that the resource is as expected.
	Expectation interface {
		id
		// Verify runs the expactation. It returns an error if the expectation is not met.
		Verify(ctx context.Context) error
	}

	// Condition is a way to define a condition for a resource. It is used to verify whether a resource is in the
	// expected state. It's meant to be used in conjunction with an Expectation. The Expectation is responsible for
	// retrieving the resource and passing it to the Condition to verify that the resource is as expected.
	Condition interface {
		id
		// IsSatisfiedBy returns `nil` if the condition is satisfied by the given resource.
		IsSatisfiedBy(ctx context.Context, resource interface{}) error
	}

	GomegaMatchable interface {
		ToGomegaMatcher() types.GomegaMatcher
	}
	GomegaEventuallySubject interface {
		GomegaMatchable
		AsGomegaSubject() func(context.Context) (interface{}, error)
	}
)
