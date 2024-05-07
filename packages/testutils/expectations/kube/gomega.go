package kube

import (
	"context"
	"fmt"

	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/onsi/gomega/gcustom"
	gomega "github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	_ types.GomegaEventuallySubject = &KubeExpectation{}
)

func (e *KubeExpectation) AsGomegaSubject() func(context.Context) (interface{}, error) {
	return func(c context.Context) (interface{}, error) {
		return e.getResults(c)
	}
}

func (e *KubeExpectation) ToGomegaMatcher() gomega.GomegaMatcher {
	return gcustom.MakeMatcher(func(resource interface{}) (bool, error) {
		list, ok := resource.([]*unstructured.Unstructured)
		if !ok {
			return false, fmt.Errorf("expected resource to be a list of unstructured.Unstructured, got %T", resource)
		}
		if err := e.verifyConditions(context.TODO(), list); err != nil {
			return false, nil
		}
		return true, nil
	})
}
