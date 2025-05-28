// Package testutils provides utilities for testing. It provides 2 main interfaces:
//
//   - Condition: a way to define conditions for an expectation
//   - Expectation: a way to define expectations for a resource
//
// This library also includes a set of predefined conditions and expectations for most common use cases.
//
// # Conditions
//
// The condition package comes with a few implementations of the Condition interface. Some of the most important constructors are:
//   - CountCondition: Creates a condition that checks if the count of a resource is equal to a given value. Works for slices and maps.
//   - JqCondition: Creates a condition that checks if the result of a jq expression is valid for the resource(s).
//   - JsonPathCondition: Creates a condition that checks if the result of a jsonpath expression is valid for the resource(s).
//   - AllCondition: Creates a condition that checks if all of the given conditions are satisfied.
//   - AnyCondition: Creates a condition that checks if any of the given conditions is satisfied.
//
// These conditions can be used to create more complex conditions. For example, the expectations.kube package has
// a predefined condition helpers to check pod readiness defined like so:
//
//	var PodReadyCondition types.Condition = conditions.All(
//	    conditions.NewKubernetesStatusCondition("Ready", true),
//	    conditions.NewKubernetesStatusCondition("Initialized", true),
//	    conditions.NewKubernetesStatusCondition("ContainersReady", true),
//	)
//
// It uses the NewKubernetesStatusCondition constructor (which itself uses the jsonpath condition constructor)
// to create a conditions that checks if the status of a pod is equal to a given value. Then it combines all of these conditions
// using the AllCondition constructor.
//
// # Expectations
//
// The epectation package comes with 4 implementations of the Expectation interface:
//   - KubernetesExpectation: an expectation for  kubernetes resources
//   - HelmExpectation: an expectation for helm releases
//   - AllExpectation: an expectation for grouping multiple expectations and checking if all of them are satisfied
//   - AnyExpectation: an expectation for grouping multiple expectations and checking if any of them is satisfied
//
// # KubernetesExpectation
//
// The KubernetesExpectation is an expectation for kubernetes resources.
// It is satisfied if the resource is present or not present in the cluster (depending on its configuration).
// The main constructor of this expectation `kube.Resource` has 3 required parameters:
//   - pattern: a regex pattern string that matches the name of the expected resource(s)
//   - namespace: the namespace of the expected resource(s). This is parameter is ignored if the resource is cluster-scoped.
//     If this resource should be matched in all namespaces, use the "*" wildcard.
//   - gvk: the group, version, kind of the expected resource(s)
//
// The expectation also accepts a list of options that can be used to configure the expectation.
// See the options section of the package documentation for more details.
//
// Because some resources are commonly expected in the cluster, this package also provides a set of predefined expectations
// and constructors for them:
//   - kube.AbsentResource: a constructor for an expectation for a resource that is not present in the cluster. ie: a resource that has been deleted or has a count of 0.
//   - kube.Pod: an expectation for a pod(s) in the cluster
//   - kube.AbsentPod: an expectation for a pod(s) that is not present in the cluster
//   - kube.Target: an expectation for a target(s) in the cluster
//   - kube.Solution: an expectation for a solution(s) in the cluster
//   - kube.Instance: an expectation for an instance(s) in the cluster
//
// # HelmExpectation
//
// The HelmExpectation is an expectation for helm releases. It is satisfied if the release is present or not present in the cluster (depending on its configuration).
// The main constructor of this expectation `helm.New` has 2 required parameters:
//   - pattern: a regex pattern string that matches the name of the expected release(s)
//   - namespace: the namespace of the expected release(s). To match releases in all namespaces, use the "*" wildcard.
//
// The expectation also accepts a list of options that can be used to configure the expectation. See the options section of the package documentation for more details.
//
// # AllExpectation
//
// The AllExpectation is an expectation for grouping multiple expectations and checking if all of them are satisfied.
// The main constructor of this expectation `expectations.All` accepts a list of expectations as parameters.
//
// # AnyExpectation
//
// The AnyExpectation is an expectation for grouping multiple expectations and checking if any of them is satisfied.
// The main constructor of this expectation `expectations.Any` accepts a list of expectations as parameters.
//
// # Examples
//
// Check if a pod named "my-pod-34jfk3-fd4k56g" in namespace "default" exists in the cluster:
//
//	exp := kube.Must(kube.Pod("my-pod-34jfk3-fd4k56g", "default"))
//	if err := exp.Verify(ctx); err != nil {
//		// expectation failed. handle error
//	}
//
// Check if there are 2 pods with prefix "my-pod-" in namespace "default" in the cluster and they are  ready:
//
//	exp := kube.Must(kube.Pod(
//		"my-pod-.*",
//		"default",
//		kube.WithListCondition(
//			conditions.Count(2)
//		),
//		kube.WithCondition(kube.PodReadyCondition), // PodReadyCondition is pre-defined in the kube package
//	))
//	if err := exp.Verify(ctx); err != nil {
//		// expectation failed. handle error
//	}
//
// Check if there are 2 pods with prefix "my-pod-" in namespace "default" in the cluster
// and that each pod is ready or initialized and each pod has a specific label
//
//	exp := kube.Must(kube.Pod(
//		"my-pod-.*",
//		"default",
//		kube.WithListCondition(
//			conditions.Count(2)
//		),
//		kube.WithCondition(conditions.All(
//			conditions.Any(
//				conditions.NewKubernetesStatusCondition("Ready", true),
//				conditions.NewKubernetesStatusCondition("Initialized", true),
//			),
//			kube.NewLabelMatchCondition("my-label", "my-value"),
//		)),
//	))
//	if err := exp.Verify(ctx); err != nil {
//		// expectation failed. handle error
//	}
package main
