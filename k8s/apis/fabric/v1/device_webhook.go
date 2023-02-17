/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var devicelog = logf.Log.WithName("device-resource")
var myDeviceClient client.Client

func (r *Device) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myDeviceClient = mgr.GetClient()

	mgr.GetFieldIndexer().IndexField(context.Background(), &Device{}, ".spec.displayName", func(rawObj client.Object) []string {
		device := rawObj.(*Device)
		return []string{device.Spec.DisplayName}
	})
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-fabric-symphony-v1-device,mutating=true,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=devices,verbs=create;update,versions=v1,name=mdevice.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Device{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Device) Default() {
	devicelog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-fabric-symphony-v1-device,mutating=false,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=devices,verbs=create;update,versions=v1,name=vdevice.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Device{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Device) ValidateCreate() error {
	devicelog.Info("validate create", "name", r.Name)

	return r.validateCreateDevice()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Device) ValidateUpdate(old runtime.Object) error {
	devicelog.Info("validate update", "name", r.Name)

	return r.validateUpdateDevice()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Device) ValidateDelete() error {
	devicelog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *Device) validateCreateDevice() error {
	var devices DeviceList
	myDeviceClient.List(context.Background(), &devices, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if len(devices.Items) != 0 {
		return fmt.Errorf("device display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}

func (r *Device) validateUpdateDevice() error {
	var devices DeviceList
	err := myDeviceClient.List(context.Background(), &devices, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return err
	}
	if !(len(devices.Items) == 0 || len(devices.Items) == 1 && devices.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return fmt.Errorf("device display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}
