/*
Copyright 2022 zoomoid.

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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kubeconfiglog = logf.Log.WithName("kubeconfig-resource")

func (r *Kubeconfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-kubeconfig-k8s-zoomoid-dev-v1alpha1-kubeconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=create;update,versions=v1alpha1,name=mkubeconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Kubeconfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Kubeconfig) Default() {
	kubeconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kubeconfig-k8s-zoomoid-dev-v1alpha1-kubeconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=create;update,versions=v1alpha1,name=vkubeconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Kubeconfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Kubeconfig) ValidateCreate() error {
	kubeconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Kubeconfig) ValidateUpdate(old runtime.Object) error {
	kubeconfiglog.Info("validate update", "name", r.Name)

	oldKubeconfig, _ := old.(*Kubeconfig)
	// check immutable fields

	if !reflect.DeepEqual(oldKubeconfig.Spec.ExistingCSR, r.Spec.ExistingCSR) {

	}
	if !reflect.DeepEqual(oldKubeconfig.Spec.CSR, r.Spec.CSR) {

	}

	return nil
}

func (r *Kubeconfig) validateExistingCSR(old runtime.Object) error {
	return nil
}

func (r *Kubeconfig) validateCSR(old runtime.Object) error {
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Kubeconfig) ValidateDelete() error {
	kubeconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
