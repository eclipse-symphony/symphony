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
