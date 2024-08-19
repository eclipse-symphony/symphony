/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"gopls-workspace/apis/dynamicclient"
	commoncontainer "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"

	configv1 "gopls-workspace/apis/config/v1"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var solutionlog = logf.Log.WithName("solution-resource")
var mySolutionReaderClient client.Reader
var projectConfig *configv1.ProjectConfig

func (r *Solution) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySolutionReaderClient = mgr.GetAPIReader()

	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	projectConfig = myConfig

	validation.SolutionInstanceLookupFunc = func(ctx context.Context, name string, namespace string) (bool, error) {
		instanceList, err := dynamicclient.ListWithLabels(validation.Instance, namespace, map[string]string{"solution": name}, 1)
		if err != nil {
			return false, err
		}
		return len(instanceList.Items) > 0, nil
	}
	validation.SolutionContainerLookupFunc = func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(validation.SolutionContainer, name, namespace)
	}
	if projectConfig.UniqueDisplayNameForSolution {
		validation.UniqueNameSolutionLookupFunc = func(ctx context.Context, displayName string, namespace string) (interface{}, error) {
			return dynamicclient.GetObjectWithUniqueName(validation.Solution, displayName, namespace)
		}
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-solution,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update,versions=v1,name=msolution.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Solution{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Solution) Default() {
	solutionlog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

	if r.Spec.RootResource != "" {
		var solutionContainer SolutionContainer
		err := mySolutionReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &solutionContainer)
		if err != nil {
			solutionlog.Error(err, "failed to get solution container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(),
				Kind:       "SolutionContainer",
				Name:       solutionContainer.Name,
				UID:        solutionContainer.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}
			if r.Labels == nil {
				r.Labels = make(map[string]string)
			}
			r.Labels["rootResource"] = r.Spec.RootResource
			if projectConfig.UniqueDisplayNameForSolution {
				r.Labels["displayName"] = r.Spec.DisplayName
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-solution,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update;delete,versions=v1,name=vsolution.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Solution{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateCreate() (admission.Warnings, error) {
	solutionlog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), solutionlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being created on namespace %s", r.Name, r.Namespace)

	return nil, r.validateCreateSolution()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	solutionlog.Info("validate update", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), solutionlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being updated on namespace %s", r.Name, r.Namespace)

	oldSolution, ok := old.(*Solution)
	if !ok {
		return nil, fmt.Errorf("expected a Target object")
	}
	return nil, r.validateUpdateSolution(oldSolution)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateDelete() (admission.Warnings, error) {
	solutionlog.Info("validate delete", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), solutionlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, r.validateDeleteSolution()
}

func (r *Solution) validateCreateSolution() error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionState()
	if err != nil {
		return err
	}
	ErrorFields := validation.ValidateCreateOrUpdate(context.TODO(), state, nil)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
}

func (r *Solution) validateUpdateSolution(old *Solution) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionState()
	if err != nil {
		return err
	}
	oldstate, err := old.ConvertSolutionState()
	if err != nil {
		return err
	}
	ErrorFields := validation.ValidateCreateOrUpdate(context.TODO(), state, oldstate)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
}

func (r *Solution) validateDeleteSolution() error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionState()
	if err != nil {
		return err
	}
	ErrorFields := validation.ValidateDelete(context.TODO(), state)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
}

func (r *Solution) ConvertSolutionState() (model.SolutionState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to solution state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.SolutionState{}, retErr
	}
	var state model.SolutionState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.SolutionState{}, retErr
	}
	return state, nil
}

func (r *SolutionContainer) Default() {
	commoncontainer.DefaultImpl(solutionlog, r)
}

func (r *SolutionContainer) ValidateCreate() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), solutionlog)

	observ_utils.EmitUserAuditsLogs(ctx, "SolutionContainer %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(solutionlog, r)
}
func (r *SolutionContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), solutionlog)

	observ_utils.EmitUserAuditsLogs(ctx, "SolutionContainer %s is being updated on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateUpdateImpl(solutionlog, r, old)
}

func (r *SolutionContainer) ValidateDelete() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionContainerOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), solutionlog)

	observ_utils.EmitUserAuditsLogs(ctx, "SolutionContainer %s is being deleted on namespace %s", r.Name, r.Namespace)

	solutionlog.Info("validate delete solution container", "name", r.Name)
	getSubResourceNums := func() (int, error) {
		var solutionList SolutionList
		err := mySolutionReaderClient.List(context.Background(), &solutionList, client.InNamespace(r.Namespace), client.MatchingLabels{"rootResource": r.Name}, client.Limit(1))
		if err != nil {
			return 0, err
		} else {
			return len(solutionList.Items), nil
		}
	}
	return commoncontainer.ValidateDeleteImpl(solutionlog, r, getSubResourceNums)
}
