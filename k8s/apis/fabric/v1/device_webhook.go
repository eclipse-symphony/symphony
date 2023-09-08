/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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

var _ webhook.Defaulter = &Device{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Device) Default() {
	devicelog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

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
