package scenarios_test

import (
	"context"
	_ "embed"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/expectations"
	"github.com/eclipse-symphony/symphony/packages/testutils/expectations/kube"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/eclipse-symphony/symphony/test/integration/lib/shell"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create/update resources for rollback testing", Ordered, func() {
	type TestCase struct {
		TargetComponents      []string
		SolutionComponents    []string
		SolutionComponentsV2  []string
		PostUpdateExpectation types.Expectation
		PostRevertExpectation types.Expectation
		TargetProperties      map[string]string
	}
	var instanceBytes []byte
	var targetBytes []byte
	var solutionBytes []byte
	var solutionBytesV2 []byte
	var targetProps map[string]string

	BeforeAll(func(ctx context.Context) {
		By("installing orchestrator in the cluster")
		shell.LocalenvCmd(ctx, "mage cluster:deploy")

		By("setting the default testing lib logger")
		logger.SetDefaultLogger(GinkgoWriter.Printf)
	})

	AfterAll(func() {
		By("uninstalling orchestrator from the cluster")
		err := shell.LocalenvCmd(context.Background(), "mage destroy all")
		Expect(err).ToNot(HaveOccurred())
	})

	JustAfterEach(func(ctx context.Context) {
		if CurrentSpecReport().Failed() {
			By("dumping cluster state")
			testhelpers.DumpClusterState(ctx)
		}
	})

	runner := func(ctx context.Context, testcase TestCase) {
		By("setting the components for the target")
		var err error

		props := targetProps
		if testcase.TargetProperties != nil {
			props = testcase.TargetProperties
		}
		// Patch the target manifest with the target options
		targetBytes, err = testhelpers.PatchTarget(defaultTargetManifest, testhelpers.TargetOptions{
			ComponentNames: testcase.TargetComponents,
			Properties:     props,
		})
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for Solution V1")
		solutionBytes, err = testhelpers.PatchSolution(defaultSolutionManifest, testhelpers.SolutionOptions{
			ComponentNames: testcase.SolutionComponents,
		})
		Expect(err).ToNot(HaveOccurred())

		By("preparing the instance bytes with a new operation id for Solution V1")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{})
		Expect(err).ToNot(HaveOccurred())

		By("deploying the Target")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", targetBytes)
		Expect(err).ToNot(HaveOccurred())

		By("deploying Solution V1")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionBytes)
		Expect(err).ToNot(HaveOccurred())

		By("deploying the Instance that references Solution V1")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for Solution V2, an invalid solution")
		solutionBytesV2, err = testhelpers.PatchSolution(defaultSolutionManifest, testhelpers.SolutionOptions{
			ComponentNames: testcase.SolutionComponentsV2,
			SolutionName:   "solution-v2",
		})
		Expect(err).ToNot(HaveOccurred())

		By("deploying Solution V2")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionBytesV2)
		Expect(err).ToNot(HaveOccurred())

		By("preparing the instance bytes with a new operation id for Solution V2")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{
			Solution: "solution-v2",
		})
		Expect(err).ToNot(HaveOccurred())

		By("updating the Instance to use Solution V2")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("verifying deployment of Instance referencing Solution V2 fails")
		err = testcase.PostUpdateExpectation.Verify(ctx)
		Expect(err).ToNot(HaveOccurred())

		By("reverting the Instance to use Solution V1")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{
			Solution: "solution",
		})
		Expect(err).ToNot(HaveOccurred())

		By("Deploying the Instance to use Solution V1 again")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("verifying deployment of Instance referencing Solution V1 succeeds")
		err = testcase.PostRevertExpectation.Verify(ctx)
		Expect(err).ToNot(HaveOccurred())
	}

	DescribeTable("fail to deploy solution v2 then rollback to v1", Ordered, runner,
		Entry("with a single component", TestCase{
			TargetComponents:     []string{"simple-chart-1"},
			SolutionComponents:   []string{"simple-chart-2"},
			SolutionComponentsV2: []string{"simple-chart-2-nonexistent"},
			PostUpdateExpectation: expectations.All(
				kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
					kube.ProvisioningFailedCondition, // and it is failed
					//jq.Equality(".status.provisioningStatus.error.details[0].details[0].code", "Update Failed"),
				)))),
			),
			PostRevertExpectation: expectations.All(
				successfullInstanceExpectation,
			),
		}),
	)
})
