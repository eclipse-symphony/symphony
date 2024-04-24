package scenarios_test

import (
	"context"
	_ "embed"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/expectations"
	"github.com/eclipse-symphony/symphony/packages/testutils/expectations/helm"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/eclipse-symphony/symphony/test/integration/lib/shell"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete", Ordered, func() {
	var instanceBytes []byte
	var targetBytes []byte
	var solutionBytes []byte
	var specTimeout = 2 * time.Minute

	type DeleteTestCase struct {
		TargetComponents        []string
		SolutionComponents      []string
		PreDeleteExpectation    types.Expectation
		UnderlyingDeleteCommand string
		OrcResourceToDelete     *[]byte
		PostDeleteExpectation   types.Expectation
	}

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

	DescribeTable("when performing create/update operations", Ordered,
		func(ctx context.Context, testcase DeleteTestCase) {
			By("setting the components for the target")
			var err error
			targetBytes, err = testhelpers.PatchTarget(defaultTargetManifest, testhelpers.TargetOptions{
				ComponentNames: testcase.TargetComponents,
			})
			Expect(err).ToNot(HaveOccurred())

			By("setting the components for the solution")
			solutionBytes, err = testhelpers.PatchSolution(defaultSolutionManifest, testhelpers.SolutionOptions{
				ComponentNames: testcase.SolutionComponents,
			})
			Expect(err).ToNot(HaveOccurred())

			By("preparing the instance bytes with a new operation id for the test")
			instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{})
			Expect(err).ToNot(HaveOccurred())

			By("deploying the instance")
			err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
			Expect(err).ToNot(HaveOccurred())

			By("deploying the target")
			err = shell.PipeInExec(ctx, "kubectl apply -f -", targetBytes)
			Expect(err).ToNot(HaveOccurred())

			By("deploying the solution")
			err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionBytes)
			Expect(err).ToNot(HaveOccurred())

			By("verifying the resources before deletion")
			err = testcase.PreDeleteExpectation.Verify(ctx)
			Expect(err).ToNot(HaveOccurred())

			By("deleting the underlying resources")
			err = shell.Exec(ctx, testcase.UnderlyingDeleteCommand)
			Expect(err).ToNot(HaveOccurred())

			By("delete the orchestrator resource")
			err = shell.PipeInExec(ctx, "kubectl delete -f -", *testcase.OrcResourceToDelete)
			Expect(err).ToNot(HaveOccurred())

			By("verifying the resources after deletion")
			err = testcase.PostDeleteExpectation.Verify(ctx)
			Expect(err).ToNot(HaveOccurred())
		},

		Entry(
			"it should delete the target when the underlying helm release is already gone", SpecTimeout(specTimeout),
			DeleteTestCase{
				TargetComponents:   []string{"simple-chart-1"},
				SolutionComponents: []string{},
				PreDeleteExpectation: expectations.All(
					successfullInstanceExpectation,
					successfullTargetExpectation,
					helm.MustNew("simple-chart-1", "azure-iot-operations", helm.WithReleaseCondition(helm.DeployedCondition)),
				),
				UnderlyingDeleteCommand: "helm uninstall simple-chart-1 -n azure-iot-operations --wait",
				OrcResourceToDelete:     &targetBytes,
				PostDeleteExpectation: expectations.All(
					successfullInstanceExpectation,
					absentTargetExpectation,
				),
			},
		),

		Entry(
			"it should delete the instance when the underlying helm release is already gone", SpecTimeout(specTimeout),
			DeleteTestCase{
				TargetComponents:   []string{},
				SolutionComponents: []string{"simple-chart-1"},
				PreDeleteExpectation: expectations.All(
					successfullInstanceExpectation,
					successfullTargetExpectation,
					helm.MustNew("simple-chart-1", "azure-iot-operations", helm.WithReleaseCondition(helm.DeployedCondition)),
				),
				UnderlyingDeleteCommand: "helm uninstall simple-chart-1 -n azure-iot-operations --wait",
				OrcResourceToDelete:     &instanceBytes,
				PostDeleteExpectation: expectations.All(
					absentInstanceExpectation,
					successfullTargetExpectation,
				),
			},
		),
		Entry(
			"it should delete the target when the underlying kubernetes resource is already gone", SpecTimeout(specTimeout),
			DeleteTestCase{
				TargetComponents:   []string{"basic-configmap-1"},
				SolutionComponents: []string{},
				PreDeleteExpectation: expectations.All(
					successfullInstanceExpectation,
					successfullTargetExpectation,
				),
				UnderlyingDeleteCommand: "kubectl delete configmap basic-configmap-1 -n azure-iot-operations",
				OrcResourceToDelete:     &targetBytes,
				PostDeleteExpectation: expectations.All(
					successfullInstanceExpectation,
					absentTargetExpectation,
				),
			},
		),

		Entry(
			"it should delete the instance when the underlying kubernetes resource is already gone", SpecTimeout(specTimeout),
			DeleteTestCase{
				TargetComponents:   []string{},
				SolutionComponents: []string{"basic-configmap-1"},
				PreDeleteExpectation: expectations.All(
					successfullInstanceExpectation,
					successfullTargetExpectation,
				),
				UnderlyingDeleteCommand: "kubectl delete configmap basic-configmap-1 -n azure-iot-operations",
				OrcResourceToDelete:     &instanceBytes,
				PostDeleteExpectation: expectations.All(
					absentInstanceExpectation,
					successfullTargetExpectation,
				),
			},
		),
	)
})
