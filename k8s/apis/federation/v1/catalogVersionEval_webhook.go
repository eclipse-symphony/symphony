package v1

import (
	"context"
	"errors"
	"fmt"
	"gopls-workspace/apis/metrics/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	configv1 "gopls-workspace/apis/config/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var catalogversionEvalExpressionLog = logf.Log.WithName("catalogversionevalexpression-resource")
var myCatalogVersionEvalExpressionClient client.Reader
var catalogversionEvalExpressionWebhookValidationMetrics *metrics.Metrics
var catalogversionEvalExpressionProjectConfig *configv1.ProjectConfig

func (r *CatalogVersionEvalExpression) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogVersionEvalExpressionClient = mgr.GetAPIReader()

	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	catalogversionEvalExpressionProjectConfig = myConfig
	// initialize the controller operation metrics
	if catalogversionEvalExpressionWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogversionEvalExpressionWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalogversionevalexpression,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogversionevalexpressions,verbs=create;update;delete,versions=v1,name=mcatalogversionevalexpression.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CatalogVersionEvalExpression{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CatalogVersionEvalExpression) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), catalogversionEvalExpressionLog)
	diagnostic.InfoWithCtx(catalogversionEvalExpressionLog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec", r.Spec, "status", r.Status)
}

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalogversionevalexpression,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogversionevalexpressions,verbs=create;update;delete,versions=v1,name=vcatalogversionevalexpression.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CatalogVersionEvalExpression{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogVersionEvalExpression) ValidateCreate() (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogVersionEvalExpression) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	// convert old object to CatalogVersionEvalExpression
	oldCatalogVersionEvalExpression, ok := old.(*CatalogVersionEvalExpression)
	if ok {
		if oldCatalogVersionEvalExpression.Status.ActionStatus.Status == SucceededActionState || oldCatalogVersionEvalExpression.Status.ActionStatus.Status == FailedActionState {
			statusStr := fmt.Sprintf("CatalogVersionEvalExpression %s under namespace %s has already reached termination status: %v", oldCatalogVersionEvalExpression.Name, oldCatalogVersionEvalExpression.Namespace, oldCatalogVersionEvalExpression.Status.ActionStatus.Status)
			resourceK8SId := r.GetNamespace() + "/" + r.GetName()
			operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Write)
			ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)
			validationError := apierrors.NewForbidden(schema.GroupResource{Group: "federation.symphony", Resource: "CatalogVersionEvalExpression"}, r.Name, errors.New("CatalogVersionEvalExpression update is not allowed when terminal state is reached"))
			diagnostic.ErrorWithCtx(catalogversionEvalExpressionLog, ctx, validationError, statusStr)
			return nil, validationError
		}
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogVersionEvalExpression) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
