package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/eclipse-symphony/symphony/k8s/apis/metrics/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type GetSubResourceNums func() (n int, err error)

var commoncontainerlog = logf.Log.WithName("commoncontainer-resource")
var cacheClient client.Client
var readerClient client.Reader
var commoncontainermetrics *metrics.Metrics

func InitCommonContainerWebHook(mgr ctrl.Manager) error {
	if commoncontainermetrics == nil {
		initmetrics, err := metrics.New()
		if err != nil {
			return err
		}
		commoncontainermetrics = initmetrics
	}
	cacheClient = mgr.GetClient()
	readerClient = mgr.GetAPIReader()
	return nil
}

func SetupWebhookWithManager(mgr ctrl.Manager, resource client.Object) error {
	mgr.GetFieldIndexer().IndexField(context.Background(), resource, ".metadata.name", func(rawObj client.Object) []string {
		return []string{rawObj.GetName()}
	})

	return ctrl.NewWebhookManagedBy(mgr).
		For(resource).
		Complete()
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CommonContainer) Default() {
	commoncontainerlog.Info("default", "name", r.Name, "kind", r.Kind)
}

func (r *CommonContainer) ValidateCreate() (admission.Warnings, error) {
	commoncontainerlog.Info("validate create", "name", r.Name, "kind", r.Kind)
	return nil, nil
}
func (r *CommonContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	commoncontainerlog.Info("validate update", "name", r.Name, "kind", r.Kind)
	return nil, nil
}

func (r *CommonContainer) ValidateDelete() (admission.Warnings, error) {
	return nil, errors.New("Not implemented")
}

func (r *CommonContainer) validateDeleteContainer() error {
	return errors.New("Not implemented")
}

func (r *CommonContainer) ValidateDeleteImpl(getSubResourceNums GetSubResourceNums) (admission.Warnings, error) {

	commoncontainerlog.Info("validate delete", "name", r.Name, "kind", r.Kind)

	validateDeleteTime := time.Now()
	validationError := r.validateDeleteContainerImpl(getSubResourceNums)
	if validationError != nil {
		commoncontainermetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.ContainerResourceType)
	} else {
		commoncontainermetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.ContainerResourceType)
	}

	return nil, validationError
}

func (r *CommonContainer) validateDeleteContainerImpl(getSubResourceNums GetSubResourceNums) error {
	itemsNum, err := getSubResourceNums()
	if err != nil {
		commoncontainerlog.Error(err, "could not list nested resources ", "name", r.Name, "kind", r.Kind)
		return apierrors.NewBadRequest(fmt.Sprintf("%s could not list nested resources for %s.", r.Kind, r.Name))
	}
	if itemsNum > 0 {
		commoncontainerlog.Error(err, "nested resources are not empty", "name", r.Name, "kind", r.Kind)
		return apierrors.NewBadRequest(fmt.Sprintf("%s nested resources with root resource '%s' are not empty", r.Kind, r.Name))
	}

	return nil
}
