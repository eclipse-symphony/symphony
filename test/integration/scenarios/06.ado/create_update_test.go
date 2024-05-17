package scenarios_test

import (
	"context"
	_ "embed"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/conditions"
	"github.com/eclipse-symphony/symphony/packages/testutils/conditions/jq"
	"github.com/eclipse-symphony/symphony/packages/testutils/expectations"
	"github.com/eclipse-symphony/symphony/packages/testutils/expectations/helm"
	"github.com/eclipse-symphony/symphony/packages/testutils/expectations/kube"
	"github.com/eclipse-symphony/symphony/packages/testutils/helpers"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/eclipse-symphony/symphony/test/integration/lib/shell"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create resources with sequential changes", Ordered, func() {
	type TestCase struct {
		TargetComponents   []string
		SolutionComponents []string
		Expectation        types.Expectation
		TargetProperties   map[string]string
		InstanceParameters map[string]interface{}
	}
	var instanceBytes []byte
	var targetBytes []byte
	var solutionBytes []byte
	var specTimeout = 120 * time.Second
	var targetProps map[string]string
	var instanceParams map[string]interface{}

	BeforeAll(func(ctx context.Context) {
		By("installing orchestrator in the cluster")
		shell.LocalenvCmd(ctx, "mage cluster:deploy")

		By("setting the default testing lib logger")
		logger.SetDefaultLogger(GinkgoWriter.Printf)
	})

	AfterAll(func() {
		By("uninstalling orchestrator from the cluster")
		//err := shell.LocalenvCmd(context.Background(), "mage destroy all")
		//Expect(err).ToNot(HaveOccurred())
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
		params := instanceParams
		if testcase.TargetProperties != nil {
			props = testcase.TargetProperties
		}

		if testcase.InstanceParameters != nil {
			params = testcase.InstanceParameters
		}
		targetBytes, err = testhelpers.PatchTarget(defaultTargetManifest, testhelpers.TargetOptions{
			ComponentNames: testcase.TargetComponents,
			Properties:     props,
		})
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for the solution")
		solutionBytes, err = testhelpers.PatchSolution(defaultSolutionManifest, testhelpers.SolutionOptions{
			ComponentNames: testcase.SolutionComponents,
		})
		Expect(err).ToNot(HaveOccurred())

		By("preparing the instance bytes with a new operation id for the test")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{
			Parameters: params,
		})
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

		err = testcase.Expectation.Verify(ctx)
		Expect(err).ToNot(HaveOccurred())
	}

	DescribeTable("when performing create/update operations", Ordered, runner,

		Entry(
			"it should deploy empty target and solution", SpecTimeout(specTimeout),
			TestCase{
				Expectation: expectations.All(
					successfullInstanceExpectation,
					successfullTargetExpectation,
				),
			},
		),

		Entry(
			"it should update the target with a simple helm chart", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1"},
				SolutionComponents: []string{},
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,      // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", "Updated"), // and the target component 'simple-chart-1' status is updated. OSS has no provisioning status yet
					)))),
					successfullInstanceExpectation,
					helm.MustNew("simple-chart-1", "azure-iot-operations", helm.WithReleaseCondition(helm.DeployedCondition)), // make sure the release is successfully deployed
				),
			},
		),

		Entry(
			"it should deploy another simple helm chart in the solution so there are 2 helm releases", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1"}, // (same as previous entry)
				SolutionComponents: []string{"simple-chart-2"},
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                       // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil), // Because nothing changed, the output should be nil
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                             // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-2", "Updated"), // and the solution component 'simple-chart-2' is created
					)))),

					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(2)), // there should be only 2 releases present
						helm.WithReleaseCondition(helm.DeployedCondition),  // all releases should have 'deployed' status
					),
					helm.MustNew("simple-chart-1", "azure-iot-operations"),
					helm.MustNew("simple-chart-2", "azure-iot-operations"),
				),
			},
		),

		Entry(
			"it should add a kubernetes config map in the solution", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1"},                      // (same as previous entry)
				SolutionComponents: []string{"simple-chart-2", "basic-configmap-1"}, //
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                       // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil), // Because the component didn't change
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                                // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-2", nil),          // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-configmap-1", "Updated"), // and the solution component 'basic-configmap-1' is created
					)))),
					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(2)), // there should be only 2 releases present
						helm.WithReleaseCondition(helm.DeployedCondition),  // all releases should have 'deployed' status
					),
					kube.Must(kube.Resource("basic-configmap-1", "azure-iot-operations", helpers.ConfigMapGVK)),
				),
			},
		),

		Entry(
			"it should add a kubernetes clusterrole in the target", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1", "basic-clusterrole"}, //
				SolutionComponents: []string{"simple-chart-2", "basic-configmap-1"}, // (same as previous entry)
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                                // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil),          // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-clusterrole", "Updated"), // and the target component 'basic-clusterrole' is created
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                          // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-2", nil),    // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-configmap-1", nil), // Because the component didn't change
					)))),
					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(2)),
						helm.WithReleaseCondition(helm.DeployedCondition),
					),
					kube.Must(kube.Resource("basic-clusterrole", "azure-iot-operations", helpers.ClusterRoleGVK)),
					kube.Must(kube.Resource("basic-configmap-1", "azure-iot-operations", helpers.ConfigMapGVK)),
				),
			},
		),

		Entry(
			"it should should just update the operation id when a no-op change is made", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1", "basic-clusterrole"}, // (same as previous entry)
				SolutionComponents: []string{"simple-chart-2", "basic-configmap-1"}, // (same as previous entry)
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                          // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil),    // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-clusterrole", nil), // Because the component didn't change
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                          // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-2", nil),    // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-configmap-1", nil), // Because the component didn't change
					)))),
					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(2)), // there should be only 2 releases present
						helm.WithReleaseCondition(helm.DeployedCondition),  // all releases should have 'deployed' status
					),
					kube.Must(kube.Resource("basic-clusterrole", "azure-iot-operations", helpers.ClusterRoleGVK)),
					kube.Must(kube.Resource("basic-configmap-1", "azure-iot-operations", helpers.GVK("", "v1", "ConfigMap"))),
				),
			},
		),

		Entry(
			"It should update remove the clusterrole from the target",
			TestCase{
				TargetComponents:   []string{"simple-chart-1"},
				SolutionComponents: []string{"simple-chart-2", "basic-configmap-1"}, // (same as previous entry)
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						////kube.OperationIdMatchCondition,                                                // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil),          // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-clusterrole", "Deleted"), // and the target component 'basic-clusterrole' is deleted
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						//kube.OperationIdMatchCondition,                                          // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-2", nil),    // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-configmap-1", nil), // Because the component didn't change
					)))),
					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(2)), // there should be only 2 releases present
						helm.WithReleaseCondition(helm.DeployedCondition),  // all releases should have 'deployed' status
					),
					kube.Must(kube.Resource("basic-configmap-1", "azure-iot-operations", helpers.GVK("", "v1", "ConfigMap"))),
					kube.Must(kube.AbsentResource("basic-clusterrole", "azure-iot-operations", helpers.ClusterRoleGVK)),
				),
			},
		),

		Entry(
			"It should update remove the simple-helmchart-2 from the solution", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1"}, // (same as previous entry)
				SolutionComponents: []string{"basic-configmap-1"},
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						//kube.OperationIdMatchCondition,                                          // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil),    // Because the component didn't change
						//kube.ProvisioningStatusComponentOutput("target.basic-clusterrole", nil), // Because it was deleted in the previous reconciliation
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						//kube.OperationIdMatchCondition,                                             // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-2", "Deleted"), // and the solution component 'simple-chart-2' is deleted
						//kube.ProvisioningStatusComponentOutput("target.basic-configmap-1", nil),    // Because the component didn't change
					)))),
					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(1)), // make sure there is only 1 release left
						helm.WithReleaseCondition(helm.DeployedCondition),  // and it is deployed
					),
					kube.Must(kube.Resource("basic-configmap-1", "azure-iot-operations", helpers.GVK("", "v1", "ConfigMap"))), // make sure the configmap still exists
					helm.MustNew("simple-chart-1", "azure-iot-operations"),                                                    // make sure simple-chart-1 is still there
					helm.MustNewAbsent("simple-chart-2", "azure-iot-operations"),                                              // make sure simple-chart-2 is gone
				),
			},
		),

		Entry(
			"It should update the simple-config-map-1 with new data", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1"},             // (same as previous entry)
				SolutionComponents: []string{"basic-configmap-1-modified"}, // (same as previous entry but with new data)
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						//kube.OperationIdMatchCondition,                                       // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-chart-1", nil), // Because the component didn't change
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						//kube.OperationIdMatchCondition,                                                // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.basic-configmap-1", "Updated"), // and the solution component 'basic-configmap-1' is updated
					)))),
					helm.MustNew("simple-chart-.*", "azure-iot-operations", // releases beginning with 'simple-chart-' in the 'azure-iot-operations' namespace
						helm.WithReleaseListCondition(conditions.Count(1)), // make sure there is only 1 release left
						helm.WithReleaseCondition(helm.DeployedCondition),  // and it is deployed
					),
					kube.Must(kube.Resource("basic-configmap-1", "azure-iot-operations", // make sure the configmap still exists
						helpers.GVK("", "v1", "ConfigMap"),
						kube.WithCondition(
							jq.Equality(".data.key", "value-modified"), // and the data is updated
						),
					)),
					helm.MustNew("simple-chart-1", "azure-iot-operations"),       // make sure simple-chart-1 is still there
					helm.MustNewAbsent("simple-chart-2", "azure-iot-operations"), // make sure simple-chart-2 is gone
				),
			},
		),

		Entry(
			"it should fail the target when component is invalid", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-chart-1-nonexistent"}, //
				SolutionComponents: []string{"basic-configmap-1-modified"}, // (same as previous entry)
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningFailedCondition, // and it is failed
						//jq.Equality(".status.provisioningStatus.error.details[0].code", "Update Failed"),
						//jq.Equality(".status.provisioningStatus.error.details[0].target", "simple-chart-1"),
					)))),
					successfullInstanceExpectation,
				),
			},
		),

		Entry(
			"it should fail the solution when component is invalid", SpecTimeout(60*time.Second),
			TestCase{
				TargetComponents:   []string{"simple-chart-1-nonexistent"}, // (same as previous entry)
				SolutionComponents: []string{"simple-chart-2-nonexistent"}, //
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningFailedCondition, // and it is failed
						//jq.Equality(".status.provisioningStatus.error.details[0].code", "Update Failed"),
						//jq.Equality(".status.provisioningStatus.error.details[0].target", "simple-chart-1"),
					)))),
					kube.Must(kube.Instance("instance", "default", kube.WithCondition(conditions.All( // make sure the instance named 'instance' is present in the 'default' namespace
						kube.ProvisioningFailedCondition, // and it is failed
						//jq.Equality(".status.provisioningStatus.error.details[0].details[0].code", "Update Failed"),
						//jq.Equality(".status.provisioningStatus.error.details[0].details[0].target", "simple-chart-2"),
					)))),
				)},
		),

		Entry(
			"it should update the target with a simple http", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-http"},
				SolutionComponents: []string{},
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningSucceededCondition, // and it is successfully provisioned
						//kube.OperationIdMatchCondition,                                          // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-http", "Updated"), // and the target component 'simple-http' status is updated
					)))),
				),
			},
		),

		Entry(
			"it should fail to update target with an invalid simple http", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents:   []string{"simple-http-invalid-url"},
				SolutionComponents: []string{},
				Expectation: expectations.All(
					kube.Must(kube.Target("target", "default", kube.WithCondition(conditions.All( // make sure the target named 'target' is present in the 'default' namespace
						kube.ProvisioningFailedCondition, // and it is failed
						//kube.OperationIdMatchCondition,                                    // and the status operation id matches the metadata operation id
						//kube.ProvisioningStatusComponentOutput("target.simple-http", nil), // and the target component 'simple-http-invalid-url' status is failed to update
						//jq.Equality(".status.provisioningStatus.error.details[0].target", "simple-http"),
						//jq.Equality(".status.provisioningStatus.error.details[0].code", "Update Failed"),
					)))),
				),
			},
		),

		// Marking as pending because even though this is correct, i.e, the order of component dependencies is respected,
		// the status probe mechanism is broken. So because symphony cannot wait for the CRD to be in an "Established" state
		// before deploying the CR, the CRD is not ready when the CR is deployed and the CR deployment fails.
		PEntry(
			"it should install the resources in the correct dependency order", SpecTimeout(specTimeout),
			TestCase{
				TargetComponents: []string{"simple-foobar", "foobar-crd"}, // explicitly orderering the CR before the CRD. This is not required but it is a good test
				Expectation: expectations.All(
					successfullInstanceExpectation,
					successfullTargetExpectation,
					kube.Must(kube.Resource("foobars.contoso.io", "azure-iot-operations", helpers.GVK("apiextensions.k8s.io", "v1", "CustomResourceDefinition"))),
					kube.Must(kube.Resource("simple-foobar", "azure-iot-operations", helpers.GVK("contoso.io", "v1", "FooBar"))),
				),
			},
		),
	)

	Context("with component constraints", func() {
		BeforeEach(func(ctx context.Context) {
			By("setting the default target props")
			targetProps = map[string]string{
				"OS": "windows",
			}
		})

		AfterEach(func(ctx context.Context) {
			By("resetting the default target props")
			targetProps = nil
		})

		DescribeTable("when performing create/update operations", Ordered, runner,
			Entry(
				"should succeed when the component is deployed to a target with matching constraint", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents: []string{"mongodb-constraint"},
					Expectation:      successfullTargetExpectation,
				},
			),
			Entry(
				"should remove config map when the component is removed from the target", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents: []string{"mongodb-constraint"},
					Expectation:      successfullTargetExpectation,
				},
			),
			Entry(
				"should fail when the component constraint references nonexistent property", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents: []string{"mongodb-constraint"},
					Expectation:      failedTargetExpectation,
					TargetProperties: map[string]string{
						"Arch": "arm",
					},
				},
			),
		)
	})
	Context("with templated expressions", func() {
		BeforeEach(func(ctx context.Context) {
			By("setting the default target props")
			targetProps = map[string]string{
				"OS":    "windows",
				"color": "blue",
			}
		})

		AfterEach(func(ctx context.Context) {
			By("resetting the default target props")
			targetProps = nil
		})

		DescribeTable("when performing create/update operations with templated expressions", Ordered, runner,
			Entry(
				"it should succeed when the component is deployed to a target with existing properties", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents: []string{"expressions-1"},
					Expectation:      successfullTargetExpectation,
				},
			),
			Entry(
				"it should fail when the component is deployed to a target with non-existing properties", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents: []string{"expressions-1-failed"},
					Expectation:      failedTargetExpectation,
				},
			),
			Entry(
				"it should succeed when solution component has a valid expression", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents:   []string{"expressions-1"},
					SolutionComponents: []string{"expressions-1-soln"},
					Expectation: expectations.All(
						successfullTargetExpectation,
						successfullInstanceExpectation,
					),
				},
			),
			Entry(
				"it should fail solution component has an invalid expression", SpecTimeout(specTimeout),
				TestCase{
					TargetComponents:   []string{"expressions-1"},
					SolutionComponents: []string{"expressions-1-soln-failed"},
					Expectation: expectations.All(
						successfullTargetExpectation,
						failedInstanceExpectation,
					),
				},
			),
		)
	})

	Context("with instance parameters", func() {
		BeforeEach(func(ctx context.Context) {
			By("setting the default instance params")
			instanceParams = map[string]interface{}{
				"database":     "mongodb",
				"database_uri": "mongodb://localhost:27017",
			}
		})

		AfterEach(func(ctx context.Context) {
			By("resetting the default instance params")
			instanceParams = nil
		})

		DescribeTable("when performing create/update operations", Ordered, runner,
			Entry(
				"should succeed when the solution component is deployed to a target with the instance parameters", SpecTimeout(specTimeout),
				TestCase{
					SolutionComponents: []string{"basic-configmap-1-params"},
					Expectation:        successfullInstanceExpectation,
				},
			),
			Entry(
				"should fail when the solution component with missing parameter is deployed to a target", SpecTimeout(specTimeout),
				TestCase{
					SolutionComponents: []string{"basic-configmap-1-params-modified"},
					Expectation:        failedInstanceExpectation,
				},
			),
		)
	})
})
