package scenarios_test

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/eclipse-symphony/symphony/packages/testutils/expectations"
	"github.com/eclipse-symphony/symphony/packages/testutils/logger"
	"github.com/eclipse-symphony/symphony/packages/testutils/types"
	"github.com/eclipse-symphony/symphony/test/integration/lib/shell"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RBAC", Ordered, func() {
	type Rbac struct {
		TargetComponents   []string
		SolutionVersionComponents []string
		InstanceScope      string
		TargetScope        string
		Expectation        types.Expectation
	}
	type HelmValues = testhelpers.HelmValues
	type Array = testhelpers.Array
	type TArray[T any] []T

	var instanceBytes []byte
	var targetBytes []byte
	var solutionversionBytes []byte
	var solutionversionContainerBytes []byte
	var specTimeout = 3 * time.Minute
	var installValues HelmValues
	var runRbacTest = func(ctx context.Context, testcase Rbac) {
		var err error

		By("deploy solutionversion container")
		solutionversionContainerBytes, err = testhelpers.PatchSolution(defaultSolutionManifest, testhelpers.ContainerOptions{})
		Expect(err).ToNot(HaveOccurred())
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionversionContainerBytes)
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for the target and scope")
		targetBytes, err = testhelpers.PatchTarget(defaultTargetManifest, testhelpers.TargetOptions{
			ComponentNames: testcase.TargetComponents,
			Scope:          testcase.TargetScope,
		})
		Expect(err).ToNot(HaveOccurred())

		By("setting the components for the solutionversion")
		solutionversionBytes, err = testhelpers.PatchSolutionVersion(defaultSolutionVersionManifest, testhelpers.SolutionVersionOptions{
			ComponentNames: testcase.SolutionVersionComponents,
		})
		Expect(err).ToNot(HaveOccurred())

		By("setting the instance scope")
		instanceBytes, err = testhelpers.PatchInstance(defaultInstanceManifest, testhelpers.InstanceOptions{
			Scope: testcase.InstanceScope,
		})
		Expect(err).ToNot(HaveOccurred())

		By("deploying the target")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", targetBytes)
		Expect(err).ToNot(HaveOccurred())

		By("deploying the solutionversion")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", solutionversionBytes)
		Expect(err).ToNot(HaveOccurred())

		By("deploying the instance")
		err = shell.PipeInExec(ctx, "kubectl apply -f -", instanceBytes)
		Expect(err).ToNot(HaveOccurred())

		By("verifying the resources")
		err = testcase.Expectation.Verify(ctx)
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeAll(func(ctx context.Context) {
		By("setting the default testing lib logger")
		logger.SetDefaultLogger(GinkgoWriter.Printf)
	})

	AfterAll(func() {
		By("uninstalling orchestrator from the cluster")
		err := shell.LocalenvCmd(context.Background(), "mage DumpSymphonyLogsForTest ginkgosuite_rbac")
		err = shell.LocalenvCmd(context.Background(), "mage Destroy all,nowait")
		Expect(err).ToNot(HaveOccurred())
	})

	JustAfterEach(func(ctx context.Context) {
		if CurrentSpecReport().Failed() {
			By("dumping cluster state")
			testhelpers.DumpClusterState(ctx)
		}
	})

	When("orchestrator is installed as cluster admin", func() {
		BeforeAll(func(ctx context.Context) {
			By("setting the install values")
			installValues = testhelpers.HelmValues{
				"rbac": testhelpers.HelmValues{
					"cluster": testhelpers.HelmValues{
						"admin": true, // Grant symphony cluster admin
					},
				},
			}

			By("installing orchestrator in the cluster")
			err := shell.LocalenvCmd(ctx, fmt.Sprintf("mage cluster:deploywithsettings '%s'", installValues.String()))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func(ctx context.Context) {
			By("removing all instances, targets and solutionversions from the cluster")
			err := testhelpers.CleanupManifests(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("when performing CRUD operations", Ordered, runRbacTest,
			Entry(
				"it succefully install in default namespace", SpecTimeout(specTimeout),
				Rbac{
					TargetComponents:   []string{"basic-clusterrole"},
					SolutionVersionComponents: []string{"simple-chart-1"},
					Expectation: expectations.All(
						successfullInstanceExpectation,
						successfullInstanceExpectation,
					),
				},
			),
		)
	})

	When("orchestrator is installed as namespace admin", func() {
		BeforeAll(func(ctx context.Context) {
			By("setting the install values")
			installValues = HelmValues{
				"rbac": HelmValues{
					"cluster": HelmValues{
						"admin": false, // Deny symphony cluster admin
					},
					"namespace": HelmValues{
						"releaseNamespaceAdmin": true, // Grant symphony namespace admin (default namespace)
					},
				},
			}

			By("installing orchestrator in the cluster")
			err := shell.LocalenvCmd(ctx, fmt.Sprintf("mage cluster:deploywithsettings '%s'", installValues.String()))
			Expect(err).ToNot(HaveOccurred())

		})

		AfterAll(func(ctx context.Context) {
			By("removing all instances, targets and solutionversions from the cluster")
			err := testhelpers.CleanupManifests(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("when performing CRUD operations", Ordered, runRbacTest,
			Entry(
				"it succefully install in default namespace", SpecTimeout(specTimeout),
				Rbac{
					TargetComponents:   []string{"mongodb-configmap"}, // Namespace level resource (configmap)
					SolutionVersionComponents: []string{"basic-configmap-1"}, // Namespace level resource (configmap)
					TargetScope:        "default",                     // Places the target component in the same namesapce as orchestrator
					InstanceScope:      "default",                     // Places the solutionversion component in the same namesapce as orchestrator
					Expectation: expectations.All(
						successfullInstanceExpectation,
						successfullInstanceExpectation,
					),
				},
			),
		)
	})

	When("orchestrator is installed with specific namespace rules", func() {
		BeforeAll(func(ctx context.Context) {
			By("setting the install values")
			installValues = HelmValues{
				"rbac": HelmValues{
					"cluster": HelmValues{
						"admin": false, // Deny symphony cluster admin
					},
					"namespace": HelmValues{
						"namespaces": HelmValues{
							"namespace-a": HelmValues{
								"rules": TArray[HelmValues]{{
									"resources": Array{"configmaps"},
									"verbs":     Array{"*"},
									"apiGroups": Array{""},
								}},
							},
						},
					},
				},
			}

			By("installing orchestrator in the cluster")
			err := shell.LocalenvCmd(ctx, fmt.Sprintf("mage cluster:deploywithsettings '%s'", installValues.String()))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func(ctx context.Context) {
			By("removing all instances, targets and solutionversions from the cluster")
			err := testhelpers.CleanupManifests(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("when performing CRUD operations", Ordered, runRbacTest,
			Entry(
				"it succefully install in default namespace", SpecTimeout(specTimeout),
				Rbac{
					TargetComponents:   []string{"mongodb-configmap"}, // Namespace level resource (configmap)
					SolutionVersionComponents: []string{"basic-configmap-1"}, // Namespace level resource (configmap)
					InstanceScope:      "namespace-a",                 // Places the solutionversion component in the allowed namespace
					TargetScope:        "namespace-a",                 // Places the target component in the allowed namespace
					Expectation: expectations.All(
						successfullInstanceExpectation,
						successfullInstanceExpectation,
					),
				},
			),
		)
	})
})
