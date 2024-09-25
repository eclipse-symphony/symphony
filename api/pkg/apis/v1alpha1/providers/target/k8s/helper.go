/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"context"
	"sort"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	apps "k8s.io/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// RsListFunc returns the ReplicaSet from the ReplicaSet namespace and the List metav1.ListOptions.
type RsListFunc func(string, metav1.ListOptions) ([]*apps.ReplicaSet, error)

// ReplicaSetsByCreationTimestamp sorts a list of ReplicaSet by creation timestamp, using their names as a tie breaker.
type ReplicaSetsByCreationTimestamp []*apps.ReplicaSet

func (o ReplicaSetsByCreationTimestamp) Len() int      { return len(o) }
func (o ReplicaSetsByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o ReplicaSetsByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// FindNewReplicaSet returns the new RS this given deployment targets (the one with the same pod template).
func FindNewReplicaSet(deployment *apps.Deployment, rsList []*apps.ReplicaSet) *apps.ReplicaSet {
	sort.Sort(ReplicaSetsByCreationTimestamp(rsList))
	for i := range rsList {
		if EqualIgnoreHash(&rsList[i].Spec.Template, &deployment.Spec.Template) {
			// In rare cases, such as after cluster upgrades, Deployment may end up with
			// having more than one new ReplicaSets that have the same template as its template,
			// see https://github.com/kubernetes/kubernetes/issues/40415
			// We deterministically choose the oldest new ReplicaSet.
			return rsList[i]
		}
	}
	// new ReplicaSet does not exist.
	return nil
}

// ListReplicaSets returns a slice of RSes the given deployment targets.
// Note that this does NOT attempt to reconcile ControllerRef (adopt/orphan),
// because only the controller itself should do that.
// However, it does filter out anything whose ControllerRef doesn't match.
func ListReplicaSets(deployment *apps.Deployment, getRSList RsListFunc) ([]*apps.ReplicaSet, error) {
	// TODO: Right now we list replica sets by their labels. We should list them by selector, i.e. the replica set's selector
	//       should be a superset of the deployment's selector, see https://github.com/kubernetes/kubernetes/issues/19830.
	namespace := deployment.Namespace
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	all, err := getRSList(namespace, options)
	if err != nil {
		return nil, err
	}
	// Only include those whose ControllerRef matches the Deployment.
	owned := make([]*apps.ReplicaSet, 0, len(all))
	for _, rs := range all {
		if metav1.IsControlledBy(rs, deployment) {
			owned = append(owned, rs)
		}
	}
	return owned, nil
}

// GetNewReplicaSet returns a replica set that matches the intent of the given deployment; get ReplicaSetList from client interface.
// Returns nil if the new replica set doesn't exist yet.
func GetNewReplicaSet(deployment *apps.Deployment, c appsclient.AppsV1Interface) (*apps.ReplicaSet, error) {
	rsList, err := ListReplicaSets(deployment, RsListFromClient(c))
	if err != nil {
		return nil, err
	}
	return FindNewReplicaSet(deployment, rsList), nil
}

// RsListFromClient returns an rsListFunc that wraps the given client.
func RsListFromClient(c appsclient.AppsV1Interface) RsListFunc {
	return func(namespace string, options metav1.ListOptions) ([]*apps.ReplicaSet, error) {
		rsList, err := c.ReplicaSets(namespace).List(context.Background(), options)
		if err != nil {
			return nil, err
		}
		var ret []*apps.ReplicaSet
		for i := range rsList.Items {
			ret = append(ret, &rsList.Items[i])
		}
		return ret, err
	}
}

// EqualIgnoreHash returns true if two given podTemplateSpec are equal, ignoring the diff in value of Labels[pod-template-hash]
// We ignore pod-template-hash because:
//  1. The hash result would be different upon podTemplateSpec API changes
//     (e.g. the addition of a new field will cause the hash code to change)
//  2. The deployment template won't have hash labels
func EqualIgnoreHash(template1, template2 *v1.PodTemplateSpec) bool {
	t1Copy := template1.DeepCopy()
	t2Copy := template2.DeepCopy()
	// Remove hash labels from template.Labels before comparing
	delete(t1Copy.Labels, apps.DefaultDeploymentUniqueLabelKey)
	delete(t2Copy.Labels, apps.DefaultDeploymentUniqueLabelKey)
	return apiequality.Semantic.DeepEqual(t1Copy, t2Copy)
}

func DeploymentReady(log logger.Logger, ctx context.Context, rs *appsv1.ReplicaSet, dep *appsv1.Deployment) bool {
	// Verify the replicaset readiness
	if !replicaSetReady(log, ctx, rs) {
		return false
	}
	// Verify the generation observed by the deployment controller matches the spec generation
	if dep.Status.ObservedGeneration != dep.ObjectMeta.Generation {
		log.InfofCtx(ctx, "  P (K8s Target): Deployment is not ready: %s/%s. observedGeneration (%d) does not match spec generation (%d).", dep.Namespace, dep.Name, dep.Status.ObservedGeneration, dep.ObjectMeta.Generation)
		return false
	}

	expectedReady := *dep.Spec.Replicas - MaxUnavailable(*dep)
	if !(rs.Status.ReadyReplicas >= expectedReady) {
		log.InfofCtx(ctx, "  P (K8s Target): Deployment is not ready: %s/%s. %d out of %d expected pods are ready", dep.Namespace, dep.Name, rs.Status.ReadyReplicas, expectedReady)
		return false
	}
	return true
}

func replicaSetReady(log logger.Logger, ctx context.Context, rs *appsv1.ReplicaSet) bool {
	// Verify the generation observed by the replicaSet controller matches the spec generation
	if rs.Status.ObservedGeneration != rs.ObjectMeta.Generation {
		log.InfofCtx(ctx, "  P (K8s Target): ReplicaSet is not ready: %s/%s. observedGeneration (%d) does not match spec generation (%d).", rs.Namespace, rs.Name, rs.Status.ObservedGeneration, rs.ObjectMeta.Generation)
		return false
	}
	return true
}

// IsRollingUpdate returns true if the strategy type is a rolling update.
func IsRollingUpdate(deployment *apps.Deployment) bool {
	return deployment.Spec.Strategy.Type == apps.RollingUpdateDeploymentStrategyType
}

// MaxUnavailable returns the maximum unavailable pods a rolling deployment can take.
func MaxUnavailable(deployment apps.Deployment) int32 {
	if !IsRollingUpdate(&deployment) || *(deployment.Spec.Replicas) == 0 {
		return int32(0)
	}
	// Error caught by validation
	_, maxUnavailable, _ := ResolveFenceposts(deployment.Spec.Strategy.RollingUpdate.MaxSurge, deployment.Spec.Strategy.RollingUpdate.MaxUnavailable, *(deployment.Spec.Replicas))
	if maxUnavailable > *deployment.Spec.Replicas {
		return *deployment.Spec.Replicas
	}
	return maxUnavailable
}

// ResolveFenceposts resolves both maxSurge and maxUnavailable. This needs to happen in one
// step. For example:
//
// 2 desired, max unavailable 1%, surge 0% - should scale old(-1), then new(+1), then old(-1), then new(+1)
// 1 desired, max unavailable 1%, surge 0% - should scale old(-1), then new(+1)
// 2 desired, max unavailable 25%, surge 1% - should scale new(+1), then old(-1), then new(+1), then old(-1)
// 1 desired, max unavailable 25%, surge 1% - should scale new(+1), then old(-1)
// 2 desired, max unavailable 0%, surge 1% - should scale new(+1), then old(-1), then new(+1), then old(-1)
// 1 desired, max unavailable 0%, surge 1% - should scale new(+1), then old(-1)
func ResolveFenceposts(maxSurge, maxUnavailable *intstrutil.IntOrString, desired int32) (int32, int32, error) {
	surge, err := intstrutil.GetValueFromIntOrPercent(intstrutil.ValueOrDefault(maxSurge, intstrutil.FromInt(0)), int(desired), true)
	if err != nil {
		return 0, 0, err
	}
	unavailable, err := intstrutil.GetValueFromIntOrPercent(intstrutil.ValueOrDefault(maxUnavailable, intstrutil.FromInt(0)), int(desired), false)
	if err != nil {
		return 0, 0, err
	}

	if surge == 0 && unavailable == 0 {
		// Validation should never allow the user to explicitly use zero values for both maxSurge
		// maxUnavailable. Due to rounding down maxUnavailable though, it may resolve to zero.
		// If both fenceposts resolve to zero, then we should set maxUnavailable to 1 on the
		// theory that surge might not work due to quota.
		unavailable = 1
	}

	return int32(surge), int32(unavailable), nil
}
