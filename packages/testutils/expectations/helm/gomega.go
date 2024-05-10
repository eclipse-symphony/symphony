package helm

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/onsi/gomega/gcustom"
	gomega "github.com/onsi/gomega/types"
	"helm.sh/helm/v3/pkg/release"
)

var _ types.GomegaEventuallySubject = &HelmExpectation{}

func (e *HelmExpectation) AsGomegaSubject() func(context.Context) (interface{}, error) {
	return func(c context.Context) (interface{}, error) {
		return e.getResults(c)
	}
}

func (e *HelmExpectation) ToGomegaMatcher() gomega.GomegaMatcher {
	return gcustom.MakeMatcher(func(resource interface{}) (bool, error) {
		releases, ok := resource.([]*release.Release)
		if !ok {
			return false, fmt.Errorf("expected resource to be a list of release.Release, got %T", resource)
		}
		if err := e.verifyConditions(context.TODO(), releases); err != nil {
			return false, nil
		}
		return true, nil
	})
}
