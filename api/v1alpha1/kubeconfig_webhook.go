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
	"context"
	"reflect"

	"github.com/zoomoid/kubeconfig-operator/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kubeconfiglog = logf.Log.WithName("kubeconfig-resource")

func (r *Kubeconfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&kubeconfigDefaulter{}).
		WithValidator(&kubeconfigValidator{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubeconfig-k8s-zoomoid-dev-v1alpha1-kubeconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=create;update,versions=v1alpha1,name=mkubeconfig.kb.io,admissionReviewVersions=v1

type kubeconfigDefaulter struct {
	client client.Client
}

var _ admission.CustomDefaulter = &kubeconfigDefaulter{}
var _ inject.Client = &kubeconfigDefaulter{}

// InjectClient injects the client into the KubeconfigValidator
func (a *kubeconfigDefaulter) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *kubeconfigDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	kubeconfig, _ := obj.(*Kubeconfig)

	kubeconfiglog.Info("default", "name", kubeconfig.Name)

	if kubeconfig.Spec.Cluster.Server == "" {
		ep, err := utils.ClusterEndpoint(ctx, r.client)
		if err != nil {
			kubeconfiglog.Error(err, "failed to get cluster endpoint from configmap in kube-public/cluster-info")
			// on retrieval error, default to localhost as a hostname and 6443
			ep = "https://localhost:6443"
		}
		kubeconfig.Spec.Cluster.Server = ep
	}

	// Default the ClusterRole ref if not specified
	if kubeconfig.Spec.RoleRef == nil {
		kubeconfig.Spec.RoleRef = &rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "cluster-admin",
			Kind:     "ClusterRole",
		}
	}

	kubeconfig.Status.defaultConditions()

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kubeconfig-k8s-zoomoid-dev-v1alpha1-kubeconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=create;update,versions=v1alpha1,name=vkubeconfig.kb.io,admissionReviewVersions=v1

type kubeconfigValidator struct {
	client client.Client
}

var _ admission.CustomValidator = &kubeconfigValidator{}

var _ inject.Client = &kubeconfigValidator{}

// InjectClient injects the client into the KubeconfigValidator
func (a *kubeconfigValidator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *kubeconfigValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	kubeconfig, _ := obj.(*Kubeconfig)
	kubeconfiglog.Info("validate create", "name", kubeconfig.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
// Update requests need to be checked for changes to the immutable fields
func (r *kubeconfigValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) error {

	oldKubeconfig, _ := old.(*Kubeconfig)
	newKubeconfig, _ := new.(*Kubeconfig)

	kubeconfiglog.Info("validate update", "name", oldKubeconfig.Name)
	var allErrs field.ErrorList
	if !reflect.DeepEqual(oldKubeconfig.Spec.ExistingCSR, newKubeconfig.Spec.ExistingCSR) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("existingCSR"), ".spec.existingCSR is immutable"))
	}
	if !reflect.DeepEqual(oldKubeconfig.Spec.CSR, newKubeconfig.Spec.CSR) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("CSR"), ".spec.csr is immutable"))
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewForbidden(schema.GroupResource{
		Group:    "kubeconfig.k8s.zoomoid.dev",
		Resource: "Kubeconfig",
	}, oldKubeconfig.Name, allErrs.ToAggregate())
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *kubeconfigValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	kubeconfig, _ := obj.(*Kubeconfig)
	kubeconfiglog.Info("validate delete", "name", kubeconfig.Name)
	return nil
}
