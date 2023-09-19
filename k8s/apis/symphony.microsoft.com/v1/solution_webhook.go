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
var solutionlog = logf.Log.WithName("solution-resource")
var mySolutionClient client.Client

func (r *Solution) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySolutionClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Solution{}, ".spec.displayName", func(rawObj client.Object) []string {
		target := rawObj.(*Solution)
		return []string{target.Spec.DisplayName}
	})
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-symphony-microsoft-com-v1-solution,mutating=true,failurePolicy=fail,sideEffects=None,groups=symphony.microsoft.com,resources=solutions,verbs=create;update,versions=v1,name=msolution.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Solution{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Solution) Default() {
	solutionlog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-symphony-microsoft-com-v1-solution,mutating=false,failurePolicy=fail,sideEffects=None,groups=symphony.microsoft.com,resources=solutions,verbs=create;update,versions=v1,name=vsolution.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Solution{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateCreate() error {
	solutionlog.Info("validate create", "name", r.Name)

	return r.validateCreateSolution()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateUpdate(old runtime.Object) error {
	solutionlog.Info("validate update", "name", r.Name)

	return r.validateUpdateSolution()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateDelete() error {
	solutionlog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *Solution) validateCreateSolution() error {
	var solutions SolutionList
	mySolutionClient.List(context.Background(), &solutions, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if len(solutions.Items) != 0 {
		return fmt.Errorf("solution display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}

func (r *Solution) validateUpdateSolution() error {
	var solutions SolutionList
	err := mySolutionClient.List(context.Background(), &solutions, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return err
	}
	if !(len(solutions.Items) == 0 || len(solutions.Items) == 1 && solutions.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return fmt.Errorf("solution display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}
