package reconcilers_test

import (
	"testing"

	"gopls-workspace/constants"
	internalTesting "gopls-workspace/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	targetOperationStartTimeKey = "target.fabric." + constants.OperationStartTimeKeyPostfix
)

func TestSuiteReconcilers(t *testing.T) {

	RegisterFailHandler(Fail)
	internalTesting.RunGinkgoSpecs(t, "Reconcilers Suite")
}
