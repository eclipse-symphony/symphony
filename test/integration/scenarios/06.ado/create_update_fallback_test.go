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
		TargetComponents            []string
		SolutionVersionComponents   []string
		SolutionVersionComponentsV2 []string
		PostUpdateExpectation       types.Expectation
		PostRevertExpectation       types.Expectation
		TargetProperties            map[string]string
	}
	var instanceBytes []byte
	var targetBytes []byte
	var solutionversionBytes []byte
	var solutionversionBytesV2 []byte
	var solutionversionContainerBytes []byte
	var targetProps map[string]string

	BeforeAll(func(ctx context.Context) {
		By("installing orchestrator in the cluster")
		shell.LocalenvCmd(ctx, "mage cluster:deploy")

		By("setting the default testing lib logger")
		logger.SetDefaultLogger(GinkgoWriter.Printf)
	})

	AfterAll(func() {
		By("uninstalling orchestrator from the cluster")
		err := shell.LocalenvCmd(context.Background(), "mage DumpSymphonyLogsForTest ginkgosuite_fallback")
		err = shell.LocalenvCmd(context.Background(), "mage Destroy all,nowait")
		Expect(err).ToNot(HaveOccurred())
	})

	JustAfterEach(func(ctx context.Context) {
		if CurrentSpecReport().Failed() {
			By("dumping cluster state")
			testhelpers.DumpClusterState(ctx)
		}
	})

	runner := func(ctx context.Context, testcase TestCase) {
		var err error

		By("deploy solutionversion container")
		solutionversionContainerBytes, err = testhelpers.PatchSolution(defaultSolutionManifest, testhelpers.ContainerOptions{})
		Expect(err).ToNot(HaveOccurred())
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionversionContainerBytes)
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for the target")
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

		By("setting the components for SolutionVersion V1")
		solutionversionBytes, err = testhelpers.PatchSolutionVersion(defaultSolutionVersionManifest, testhelpers.SolutionVersionOptions{
			ComponentNames: testcase.SolutionVersionComponents,
		})
		Expect(err).ToNot(HaveOccurred())

		By("preparing the instance bytes with a new operation id for SolutionVersion V1")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{})
		Expect(err).ToNot(HaveOccurred())

		By("deploying the Target")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", targetBytes)
		Expect(err).ToNot(HaveOccurred())

		By("deploying SolutionVersion V1")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionversionBytes)
		Expect(err).ToNot(HaveOccurred())

		By("deploying the Instance that references SolutionVersion V1")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for SolutionVersion V2, an invalid solutionversion")
		solutionversionBytesV2, err = testhelpers.PatchSolutionVersion(defaultSolutionVersionManifest, testhelpers.SolutionVersionOptions{
			ComponentNames:      testcase.SolutionVersionComponentsV2,
			SolutionVersionName: "solution-v-version2",
		})
		Expect(err).ToNot(HaveOccurred())

		By("deploying SolutionVersion V2")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionversionBytesV2)
		Expect(err).ToNot(HaveOccurred())

		By("preparing the instance bytes with a new operation id for SolutionVersion V2")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{
			SolutionVersion: "solution:version2",
		})
		Expect(err).ToNot(HaveOccurred())

		By("updating the Instance to use SolutionVersion V2")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("verifying deployment of Instance referencing SolutionVersion V2 fails")
		err = testcase.PostUpdateExpectation.Verify(ctx)
		Expect(err).ToNot(HaveOccurred())

		By("reverting the Instance to use SolutionVersion V1")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{
			SolutionVersion: "solution:version1",
		})
		Expect(err).ToNot(HaveOccurred())

		By("Deploying the Instance to use SolutionVersion V1 again")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("verifying deployment of Instance referencing SolutionVersion V1 succeeds")
		err = testcase.PostRevertExpectation.Verify(ctx)
		Expect(err).ToNot(HaveOccurred())
	}

	DescribeTable("fail to deploy solutionversion v2 then rollback to v1", Ordered, runner,
		Entry("with a single component", TestCase{
			TargetComponents:            []string{"simple-chart-1"},
			SolutionVersionComponents:   []string{"simple-chart-2"},
			SolutionVersionComponentsV2: []string{"simple-chart-2-nonexistent"},
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
