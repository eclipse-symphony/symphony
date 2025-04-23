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
	"gopls-workspace/utils/diagnostic"

	configv1 "gopls-workspace/apis/config/v1"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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

var (
	solutionContainerMaxNameLength = 61
	solutionContainerMinNameLength = 3
	solutionlog                    = logf.Log.WithName("solution-resource")
	mySolutionReaderClient         client.Reader
	projectConfig                  *configv1.ProjectConfig
	solutionValidator              validation.SolutionValidator
)

func (r *Solution) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySolutionReaderClient = mgr.GetAPIReader()

	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	projectConfig = myConfig

	// Load validator functions
	solutionInstanceLookupFunc := func(ctx context.Context, name string, namespace string, solutionUid string) (bool, error) {
		instanceList, err := dynamicclient.ListWithLabels(ctx, validation.Instance, namespace, map[string]string{api_constants.SolutionUid: solutionUid}, 1)
		if err != nil {
			return false, err
		}
		// use uid label first and then name label
		if len(instanceList.Items) > 0 {
			diagnostic.InfoWithCtx(solutionlog, ctx, "solution look up instance using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "solution (%s) in namespace (%s) look up instance using UID ", r.Name, r.Namespace)
			return len(instanceList.Items) > 0, nil
		}
		if len(name) < api_constants.LabelLengthUpperLimit {
			instanceList, err = dynamicclient.ListWithLabels(ctx, validation.Instance, namespace, map[string]string{api_constants.Solution: name}, 1)
			if err != nil {
				return false, err
			}
			if len(instanceList.Items) > 0 {
				diagnostic.InfoWithCtx(solutionlog, ctx, "solution look up instance using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "solution (%s) in namespace (%s) look up instance using NAME ", r.Name, r.Namespace)
				return len(instanceList.Items) > 0, nil
			}
		}
		return false, nil
	}
	solutionContainerLookupFunc := func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(ctx, validation.SolutionContainer, name, namespace)
	}

	uniqueNameSolutionLookupFunc := func(ctx context.Context, displayName string, namespace string) (interface{}, error) {
		return dynamicclient.GetObjectWithUniqueName(ctx, validation.Solution, displayName, namespace)
	}
	if projectConfig.UniqueDisplayNameForSolution {
		solutionValidator = validation.NewSolutionValidator(solutionInstanceLookupFunc, solutionContainerLookupFunc, uniqueNameSolutionLookupFunc)
	} else {
		solutionValidator = validation.NewSolutionValidator(solutionInstanceLookupFunc, solutionContainerLookupFunc, nil)
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
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), solutionlog)
	diagnostic.InfoWithCtx(solutionlog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec", r.Spec, "status", r.Status)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

	if r.Spec.RootResource != "" {
		var solutionContainer SolutionContainer
		err := mySolutionReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &solutionContainer)
		if err != nil {
			diagnostic.ErrorWithCtx(solutionlog, ctx, err, "failed to get solution container", "name", r.Spec.RootResource)
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

			// Remove api_constants.RootResource from r.Labels if it exists
			if _, exists := r.Labels[api_constants.RootResource]; exists {
				delete(r.Labels, api_constants.RootResource)
			}
			var solutionContainer SolutionContainer
			err := mySolutionReaderClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.RootResource), Namespace: r.Namespace}, &solutionContainer)
			if err != nil {
				diagnostic.ErrorWithCtx(solutionlog, ctx, err, "failed to get solutionContainer", "name", r.Name, "namespace", r.Namespace)
			}
			r.Labels[api_constants.RootResourceUid] = string(solutionContainer.UID)
			if projectConfig.UniqueDisplayNameForSolution {
				r.Labels[api_constants.DisplayName] = utils.ConvertStringToValidLabel(r.Spec.DisplayName)
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-solution,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update;delete,versions=v1,name=vsolution.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Solution{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionReaderClient, context.TODO(), solutionlog)

	diagnostic.InfoWithCtx(solutionlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Solution %s is being created on namespace %s", r.Name, r.Namespace)

	return nil, r.validateCreateSolution(ctx)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionReaderClient, context.TODO(), solutionlog)

	diagnostic.InfoWithCtx(solutionlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Solution %s is being updated on namespace %s", r.Name, r.Namespace)

	oldSolution, ok := old.(*Solution)
	if !ok {
		err := fmt.Errorf("expected a Solution object")
		diagnostic.ErrorWithCtx(solutionlog, ctx, err, "failed to convert old object to Solution")
		return nil, err
	}
	return nil, r.validateUpdateSolution(ctx, oldSolution)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionReaderClient, context.TODO(), solutionlog)

	diagnostic.InfoWithCtx(solutionlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Solution %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, r.validateDeleteSolution(ctx)
}

func (r *Solution) validateCreateSolution(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate create solution - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := solutionValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate create solution", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Solution) validateUpdateSolution(ctx context.Context, old *Solution) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate update solution - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	oldstate, err := old.ConvertSolutionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate update solution - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := solutionValidator.ValidateCreateOrUpdate(context.TODO(), state, oldstate)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate update solution", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Solution) validateDeleteSolution(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate delete solution - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := solutionValidator.ValidateDelete(ctx, state)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(solutionlog, ctx, err, "validate delete solution", "name", r.Name, "namespace", r.Namespace)
	return err
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
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), solutionlog)
	commoncontainer.DefaultImpl(solutionlog, ctx, r)
}

func (r *SolutionContainer) ValidateCreate() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionReaderClient, context.TODO(), solutionlog)

	diagnostic.InfoWithCtx(solutionlog, ctx, "validate create solution container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "SolutionContainer %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(solutionlog, ctx, r, solutionContainerMinNameLength, solutionContainerMaxNameLength)
}
func (r *SolutionContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionReaderClient, context.TODO(), solutionlog)

	diagnostic.InfoWithCtx(solutionlog, ctx, "validate update solution container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "SolutionContainer %s is being updated on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateUpdateImpl(solutionlog, ctx, r, old)
}

func (r *SolutionContainer) ValidateDelete() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionContainerOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionReaderClient, context.TODO(), solutionlog)

	diagnostic.InfoWithCtx(solutionlog, ctx, "validate delete solution container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "SolutionContainer %s is being deleted on namespace %s", r.Name, r.Namespace)

	getSubResourceNums := func() (int, error) {
		var solutionList SolutionList
		err := mySolutionReaderClient.List(context.Background(), &solutionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResourceUid: string(r.UID)}, client.Limit(1))
		if err != nil {
			diagnostic.ErrorWithCtx(solutionlog, ctx, err, "failed to list solutions", "name", r.Name, "namespace", r.Namespace)
			return 0, err
		}

		if len(solutionList.Items) > 0 {
			diagnostic.InfoWithCtx(solutionlog, ctx, "solutioncontainer look up solution using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "solutioncontainer (%s) in namespace (%s) look up solution using UID ", r.Name, r.Namespace)
			return len(solutionList.Items), nil
		}

		if len(r.Name) < api_constants.LabelLengthUpperLimit {
			err = mySolutionReaderClient.List(context.Background(), &solutionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
			if err != nil {
				diagnostic.ErrorWithCtx(solutionlog, ctx, err, "failed to list solutions", "name", r.Name, "namespace", r.Namespace)
				return 0, err
			}
			if len(solutionList.Items) > 0 {
				diagnostic.InfoWithCtx(solutionlog, ctx, "solutioncontainer look up solution using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "solutioncontainer (%s) in namespace (%s) look up solution using NAME ", r.Name, r.Namespace)
				return len(solutionList.Items), nil
			}
		}
		return 0, nil
	}
	return commoncontainer.ValidateDeleteImpl(solutionlog, ctx, r, getSubResourceNums)
}
