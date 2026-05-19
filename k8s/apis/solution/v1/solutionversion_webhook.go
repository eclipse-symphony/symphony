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
	solutionversionContainerMaxNameLength = 61
	solutionversionContainerMinNameLength = 1
	solutionversionlog                    = logf.Log.WithName("solutionversion-resource")
	mySolutionVersionReaderClient         client.Reader
	projectConfig                  *configv1.ProjectConfig
	solutionversionValidator              validation.SolutionVersionValidator
)

func (r *SolutionVersion) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySolutionVersionReaderClient = mgr.GetAPIReader()

	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	projectConfig = myConfig

	// Load validator functions
	solutionversionInstanceLookupFunc := func(ctx context.Context, name string, namespace string, solutionversionUid string) (bool, error) {
		instanceList, err := dynamicclient.ListWithLabels(ctx, validation.Instance, namespace, map[string]string{api_constants.SolutionVersionUid: solutionversionUid}, 1)
		if err != nil {
			return false, err
		}
		// use uid label first and then name label
		if len(instanceList.Items) > 0 {
			diagnostic.InfoWithCtx(solutionversionlog, ctx, "solutionversion look up instance using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "solutionversion (%s) in namespace (%s) look up instance using UID ", r.Name, r.Namespace)
			return len(instanceList.Items) > 0, nil
		}
		if len(name) < api_constants.LabelLengthUpperLimit {
			instanceList, err = dynamicclient.ListWithLabels(ctx, validation.Instance, namespace, map[string]string{api_constants.SolutionVersion: name}, 1)
			if err != nil {
				return false, err
			}
			if len(instanceList.Items) > 0 {
				diagnostic.InfoWithCtx(solutionversionlog, ctx, "solutionversion look up instance using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "solutionversion (%s) in namespace (%s) look up instance using NAME ", r.Name, r.Namespace)
				return len(instanceList.Items) > 0, nil
			}
		}
		return false, nil
	}
	solutionversionContainerLookupFunc := func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(ctx, validation.Solution, name, namespace)
	}

	uniqueNameSolutionVersionLookupFunc := func(ctx context.Context, displayName string, namespace string) (interface{}, error) {
		return dynamicclient.GetObjectWithUniqueName(ctx, validation.SolutionVersion, displayName, namespace)
	}
	if projectConfig.UniqueDisplayNameForSolutionVersion {
		solutionversionValidator = validation.NewSolutionVersionValidator(solutionversionInstanceLookupFunc, solutionversionContainerLookupFunc, uniqueNameSolutionVersionLookupFunc)
	} else {
		solutionversionValidator = validation.NewSolutionVersionValidator(solutionversionInstanceLookupFunc, solutionversionContainerLookupFunc, nil)
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solutionversion-symphony-v1-solutionversion,mutating=true,failurePolicy=fail,sideEffects=None,groups=solutionversion.symphony,resources=solutionversions,verbs=create;update,versions=v1,name=msolutionversion.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &SolutionVersion{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SolutionVersion) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), solutionversionlog)
	diagnostic.InfoWithCtx(solutionversionlog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "status", r.Status)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

	if r.Spec.RootResource != "" {
		var solutionversionContainer Solution
		err := mySolutionVersionReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &solutionversionContainer)
		if err != nil {
			diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "failed to get solutionversion container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(),
				Kind:       "Solution",
				Name:       solutionversionContainer.Name,
				UID:        solutionversionContainer.UID,
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
			var solutionversionContainer Solution
			err := mySolutionVersionReaderClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.RootResource), Namespace: r.Namespace}, &solutionversionContainer)
			if err != nil {
				diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "failed to get solutionversionContainer", "name", r.Name, "namespace", r.Namespace)
			}
			r.Labels[api_constants.RootResourceUid] = string(solutionversionContainer.UID)
			if projectConfig.UniqueDisplayNameForSolutionVersion {
				r.Labels[api_constants.DisplayName] = utils.ConvertStringToValidLabel(r.Spec.DisplayName)
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solutionversion-symphony-v1-solutionversion,mutating=false,failurePolicy=fail,sideEffects=None,groups=solutionversion.symphony,resources=solutionversions,verbs=create;update;delete,versions=v1,name=vsolutionversion.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SolutionVersion{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SolutionVersion) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionVersionReaderClient, context.TODO(), solutionversionlog)

	diagnostic.InfoWithCtx(solutionversionlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "SolutionVersion %s is being created on namespace %s", r.Name, r.Namespace)

	return nil, r.validateCreateSolutionVersion(ctx)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SolutionVersion) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionVersionReaderClient, context.TODO(), solutionversionlog)

	diagnostic.InfoWithCtx(solutionversionlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "SolutionVersion %s is being updated on namespace %s", r.Name, r.Namespace)

	oldSolutionVersion, ok := old.(*SolutionVersion)
	if !ok {
		err := fmt.Errorf("expected a SolutionVersion object")
		diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "failed to convert old object to SolutionVersion")
		return nil, err
	}
	return nil, r.validateUpdateSolutionVersion(ctx, oldSolutionVersion)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SolutionVersion) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionVersionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionVersionReaderClient, context.TODO(), solutionversionlog)

	diagnostic.InfoWithCtx(solutionversionlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "SolutionVersion %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, r.validateDeleteSolutionVersion(ctx)
}

func (r *SolutionVersion) validateCreateSolutionVersion(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate create solutionversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := solutionversionValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solutionversion.symphony", Kind: "SolutionVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate create solutionversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *SolutionVersion) validateUpdateSolutionVersion(ctx context.Context, old *SolutionVersion) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate update solutionversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	oldstate, err := old.ConvertSolutionVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate update solutionversion - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := solutionversionValidator.ValidateCreateOrUpdate(context.TODO(), state, oldstate)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solutionversion.symphony", Kind: "SolutionVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate update solutionversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *SolutionVersion) validateDeleteSolutionVersion(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertSolutionVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate delete solutionversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := solutionversionValidator.ValidateDelete(ctx, state)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solutionversion.symphony", Kind: "SolutionVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "validate delete solutionversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *SolutionVersion) ConvertSolutionVersionState() (model.SolutionVersionState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "solutionversion.symphony", Kind: "SolutionVersion"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to solutionversion state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.SolutionVersionState{}, retErr
	}
	var state model.SolutionVersionState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.SolutionVersionState{}, retErr
	}
	return state, nil
}

func (r *Solution) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), solutionversionlog)
	commoncontainer.DefaultImpl(solutionversionlog, ctx, r)
}

func (r *Solution) ValidateCreate() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionVersionReaderClient, context.TODO(), solutionversionlog)

	diagnostic.InfoWithCtx(solutionversionlog, ctx, "validate create solutionversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Solution %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(solutionversionlog, ctx, r, solutionversionContainerMinNameLength, solutionversionContainerMaxNameLength)
}
func (r *Solution) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionVersionReaderClient, context.TODO(), solutionversionlog)

	diagnostic.InfoWithCtx(solutionversionlog, ctx, "validate update solutionversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Solution %s is being updated on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateUpdateImpl(solutionversionlog, ctx, r, old)
}

func (r *Solution) ValidateDelete() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.SolutionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, mySolutionVersionReaderClient, context.TODO(), solutionversionlog)

	diagnostic.InfoWithCtx(solutionversionlog, ctx, "validate delete solutionversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Solution %s is being deleted on namespace %s", r.Name, r.Namespace)

	getSubResourceNums := func() (int, error) {
		var solutionversionList SolutionVersionList
		err := mySolutionVersionReaderClient.List(context.Background(), &solutionversionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResourceUid: string(r.UID)}, client.Limit(1))
		if err != nil {
			diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "failed to list solutionversions", "name", r.Name, "namespace", r.Namespace)
			return 0, err
		}

		if len(solutionversionList.Items) > 0 {
			diagnostic.InfoWithCtx(solutionversionlog, ctx, "solution look up solutionversion using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "solution (%s) in namespace (%s) look up solutionversion using UID ", r.Name, r.Namespace)
			return len(solutionversionList.Items), nil
		}

		if len(r.Name) < api_constants.LabelLengthUpperLimit {
			err = mySolutionVersionReaderClient.List(context.Background(), &solutionversionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
			if err != nil {
				diagnostic.ErrorWithCtx(solutionversionlog, ctx, err, "failed to list solutionversions", "name", r.Name, "namespace", r.Namespace)
				return 0, err
			}
			if len(solutionversionList.Items) > 0 {
				diagnostic.InfoWithCtx(solutionversionlog, ctx, "solution look up solutionversion using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "solution (%s) in namespace (%s) look up solutionversion using NAME ", r.Name, r.Namespace)
				return len(solutionversionList.Items), nil
			}
		}
		return 0, nil
	}
	return commoncontainer.ValidateDeleteImpl(solutionversionlog, ctx, r, getSubResourceNums)
}
