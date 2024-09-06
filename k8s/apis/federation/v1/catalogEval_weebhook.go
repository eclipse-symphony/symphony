package v1

import (
	"errors"
	"gopls-workspace/apis/metrics/v1"
	"gopls-workspace/configutils"

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
var catalogEvalExpressionLog = logf.Log.WithName("catalogevalexpression-resource")
var myCatalogEvalExpressionClient client.Reader
var catalogEvalExpressionWebhookValidationMetrics *metrics.Metrics
var catalogEvalExpressionProjectConfig *configv1.ProjectConfig

func (r *CatalogEvalExpression) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogEvalExpressionClient = mgr.GetAPIReader()

	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	catalogEvalExpressionProjectConfig = myConfig
	// initialize the controller operation metrics
	if catalogEvalExpressionWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogEvalExpressionWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalogevalexpression,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogevalexpressions,verbs=create;update;delete,versions=v1,name=mcatalogevalexpression.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CatalogEvalExpression{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CatalogEvalExpression) Default() {
	catalogEvalExpressionLog.Info("default", "name", r.GetName(), "kind", r.GetObjectKind())
}

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalogevalexpression,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogevalexpressions,verbs=create;update;delete,versions=v1,name=vcatalogevalexpression.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CatalogEvalExpression{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogEvalExpression) ValidateCreate() (admission.Warnings, error) {
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogEvalExpression) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	// convert old object to CatalogEvalExpression
	oldCatalogEvalExpression, ok := old.(*CatalogEvalExpression)
	if ok {
		if oldCatalogEvalExpression.Status.ActionStatus.Status == SucceededActionState || oldCatalogEvalExpression.Status.ActionStatus.Status == FailedActionState {
			validationError := apierrors.NewForbidden(schema.GroupResource{Group: "federation.symphony", Resource: "CatalogEvalExpression"}, r.Name, errors.New("CatalogEvalExpression update is not allowed when terminal state is reached"))
			return nil, validationError
		}
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogEvalExpression) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
