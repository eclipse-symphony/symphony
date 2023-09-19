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
var instancelog = logf.Log.WithName("instance-resource")
var myInstanceClient client.Client

func (r *Instance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myInstanceClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, ".spec.displayName", func(rawObj client.Object) []string {
		target := rawObj.(*Instance)
		return []string{target.Spec.DisplayName}
	})
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, "spec.solution", func(rawObj client.Object) []string {
		target := rawObj.(*Instance)
		return []string{target.Spec.Solution}
	})
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update,versions=v1,name=minstance.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Instance{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Instance) Default() {
	instancelog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update,versions=v1,name=vinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Instance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateCreate() error {
	instancelog.Info("validate create", "name", r.Name)

	return r.validateCreateInstance()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateUpdate(old runtime.Object) error {
	instancelog.Info("validate update", "name", r.Name)

	return r.validateUpdateInstance()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateDelete() error {
	instancelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Instance) validateCreateInstance() error {
	var instances InstanceList
	myInstanceClient.List(context.Background(), &instances, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if len(instances.Items) != 0 {
		return fmt.Errorf("instance display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}

func (r *Instance) validateUpdateInstance() error {
	var instances InstanceList
	err := myInstanceClient.List(context.Background(), &instances, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return err
	}
	if !(len(instances.Items) == 0 || len(instances.Items) == 1 && instances.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return fmt.Errorf("instance display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}
