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

package controllers

import (
	"context"
	"fmt"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/tools/record"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	"github.com/zoomoid/kubeconfig-operator/controllers/phases"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeconfigReconciler reconciles a Kubeconfig object
type KubeconfigReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *KubeconfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	kubeconfig := &kubeconfigv1alpha1.Kubeconfig{}
	err := r.Get(ctx, req.NamespacedName, kubeconfig)

	if err != nil {
		klog.V(1).ErrorS(err, "failed to get kubeconfig object")
		if apierrors.IsNotFound(err) {
			// this is a delete request, kubernetes will garbage collect all other objects
			// with ownership from this controller
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	klog.V(2).InfoS("Reconciling kubeconfig", "name", kubeconfig.Name)

	terminal := isInTerminalCondition(kubeconfig)
	if terminal {
		klog.V(2).InfoS("Kubeconfig is done, skipping reconciliation", "name", kubeconfig.Name)
		return ctrl.Result{}, nil
	}

	var userSecretName types.NamespacedName
	userSecret := &corev1.Secret{}
	if kubeconfig.Spec.ExistingCSR != nil {
		userSecretName = types.NamespacedName{
			Namespace: kubeconfig.Spec.ExistingCSR.Namespace,
			Name:      kubeconfig.Spec.ExistingCSR.Name,
		}
		err = r.Get(ctx, userSecretName, userSecret)
		if apierrors.IsNotFound(err) {
			klog.Error(err, "failed to get predefined secret for CSR data, waiting for appearance")
			r.Recorder.Eventf(kubeconfig, "Warning", "NotFound", "Predefined secret for CSR could not be found, %v", err)
			return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Minute}, err
		}
		kubeconfig.Status.UserSecret = kubeconfigv1alpha1.SecretRef{
			Namespace: userSecret.Namespace,
			Name:      userSecret.Name,
		}
		// make the controller the owner of the resource
		ctrl.SetControllerReference(kubeconfig, userSecret, r.Scheme)
	} else {
		userSecretName = types.NamespacedName{
			Name: fmt.Sprintf("%s-client-cert", kubeconfig.Name),
			// since secrets are namespaced, we need to create it somewhere, and we put this secret into the operator's namespace
			Namespace: "kubeconfig-operator-system",
		}
		err = r.Get(ctx, userSecretName, userSecret)
		if apierrors.IsNotFound(err) {
			userSecret = r.createUserSecret(kubeconfig, userSecretName)
			err = r.Create(ctx, userSecret)
			if err != nil {
				klog.Error(err)
				r.Recorder.Eventf(kubeconfig, "Warning", "UserSecretFailed", "Failed to create user secret, %v", err)
				meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
					Type:    kubeconfigv1alpha1.ConditionTypeUserSecretCreated,
					Reason:  "UserSecretCreateFailed",
					Message: fmt.Sprintf("Failed to create user secret, %v", err),
					Status:  metav1.ConditionFalse,
				})
				kubeconfig.Status.Status = phases.PhaseFailed
				r.Status().Update(ctx, kubeconfig)
				return ctrl.Result{}, err
			}
			kubeconfig.Status.UserSecret = kubeconfigv1alpha1.SecretRef{
				Namespace: userSecret.Namespace,
				Name:      userSecret.Name,
			}
			meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
				Type:    kubeconfigv1alpha1.ConditionTypeUserSecretCreated,
				Reason:  "UserSecretCreated",
				Message: "Created user secret",
				Status:  metav1.ConditionTrue,
			})
			klog.V(2).InfoS("Created user secret", "namespace", userSecret.Namespace, "name", userSecret.Name)
		} else if err != nil {
			return ctrl.Result{}, err
		}
	}

	csr := &certificatesv1.CertificateSigningRequest{}
	err = r.Get(ctx, types.NamespacedName{Name: kubeconfig.Name}, csr)
	if apierrors.IsNotFound(err) {
		// condition is either false or unknown, either way create a fresh CSR, create a fresh CSR
		klog.V(2).InfoS("Creating a fresh CSR for kubeconfig", "name", kubeconfig.Name)
		r.Recorder.Event(kubeconfig, "Normal", "Generating", "Generating CSR for kubeconfig")
		keyBuffer, csrBuffer, err := r.createCSR(kubeconfig)
		if err != nil {
			// append failure condition to Kubeconfig object
			meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
				Type:    kubeconfigv1alpha1.ConditionTypeCSRCreated,
				Reason:  "CsrCreateFailed",
				Message: fmt.Sprintf("Failed to generate private key and certificate signing request, %v", err),
				Status:  metav1.ConditionFalse,
			})
			kubeconfig.Status.Status = phases.PhaseFailed
			r.Recorder.Eventf(kubeconfig, "Warning", "CsrFailed", "Failed to generate private key and certificate signing request, %v", err)
			klog.ErrorS(err, "Failed to create CSR for kubeconfig", "name", kubeconfig.Name)
			return ctrl.Result{}, nil
		}

		userSecret.Data[CertificateSecretPrivKeyKey] = keyBuffer.Bytes()
		userSecret.Data[CertificateSecretCSRKey] = csrBuffer.Bytes()

		err = r.Update(ctx, userSecret)
		if err != nil {
			klog.ErrorS(err, "Failed to update user secret with private key and CSR buffers", "namespace", userSecret.Namespace, "name", userSecret.Name)
			r.Recorder.Eventf(kubeconfig, "Warning", "UserSecretUpdated", "Failed to update user secret with private key and CSR, %v", err)
			return ctrl.Result{}, err
		}

		klog.V(2).InfoS("Updated user secret", "namespace", userSecret.Namespace, "name", userSecret.Name)
		r.Recorder.Event(kubeconfig, "Normal", "UserSecretUpdated", "Added private key and CSR to user secret")

		// Create fresh CSR and a secret keeping track of the private/public key and the CSR
		csr := r.createCsr(kubeconfig, csrBuffer)
		err = r.Create(ctx, csr)
		if err != nil {
			klog.Error(err)
			r.Recorder.Eventf(kubeconfig, "Warning", "CsrFailed", "Failed to create CSR, %v", err)
			return ctrl.Result{}, err
		}
		kubeconfig.Status.Csr = kubeconfigv1alpha1.CsrRef{
			Name: csr.Name,
		}
		meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
			Type:    kubeconfigv1alpha1.ConditionTypeCSRCreated,
			Reason:  "CsrCreated",
			Message: "Created CSR for kubeconfig request",
			Status:  metav1.ConditionTrue,
		})
		kubeconfig.Status.Status = phases.PhaseAwaitingApproval
		r.Recorder.Eventf(kubeconfig, "Normal", "Created", "Created user secret and CSR for kubeconfig")
		klog.V(0).InfoS("Exiting early, created CSR, waiting for next reconciliation", "name", kubeconfig.Name)
		r.Status().Update(ctx, kubeconfig)
		return ctrl.Result{}, nil
	} else if err != nil {
		klog.ErrorS(err, "Failed to get CSR from apiserver")
		return ctrl.Result{}, err
	}

	approved, denied, failed := getCertApprovalCondition(csr.Status.Conditions)
	if denied || failed {
		// updated approval state for this reonciler run
		klog.Errorf("CSR was denied or failed", "failed", failed, "denied", denied)
		meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
			Type:    kubeconfigv1alpha1.ConditionTypeCSRApproved,
			Status:  metav1.ConditionFalse,
			Reason:  "CsrApproved",
			Message: "CSR for the kubeconfig was denied or failed",
		})
		meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
			Type:    kubeconfigv1alpha1.ConditionTypeKubeconfigFinished,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: "Kubeconfig creation failed in CSR stage",
		})
		kubeconfig.Status.Status = phases.PhaseFailed
		r.Status().Update(ctx, kubeconfig)
		return ctrl.Result{}, nil
	}
	if !approved {
		// CSR is still pending somehow
		return ctrl.Result{}, nil
	}

	// updated approval state for this reonciler run
	meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
		Type:    kubeconfigv1alpha1.ConditionTypeCSRApproved,
		Status:  metav1.ConditionTrue,
		Reason:  "CsrApproved",
		Message: "CSR for the kubeconfig was approved",
	})

	cert := csr.Status.Certificate
	if len(cert) == 0 {
		klog.V(0).InfoS("Certificate is empty, requeuing kubeconfig reconciliation", "name", kubeconfig.Name)
		r.Status().Update(ctx, kubeconfig)
		return ctrl.Result{RequeueAfter: 15 * time.Second, Requeue: true}, nil
	}
	// Upsert secret with certificate
	userSecret.Data[CertificateSecretCertKey] = cert

	cfg, err := r.createKubeconfig(ctx, kubeconfig, userSecret)
	if err != nil {
		klog.ErrorS(err, "failed to template kubeconfig")
		r.Recorder.Eventf(kubeconfig, "Warning", "KubeconfigSecretFailed", "Failed to template kubeconfig, %v", err)
		return ctrl.Result{}, err
	}
	kubeconfig.Status.Kubeconfig = string(cfg)
	userSecret.Data[KubeconfigKey] = cfg
	err = r.Update(ctx, userSecret)
	if err != nil {
		// an error updating the user secret should NOT mark the kubeconfig as terminally failed,
		// instead, it should requeue the request at a later time (i.e. one minute)
		klog.ErrorS(err, "failed to update user secret", "namespace", userSecret.Namespace, "name", userSecret.Name)
		return ctrl.Result{}, err
	}
	klog.InfoS("Updated user secret secret", "namespace", userSecret.Namespace, "name", userSecret.Name)

	// update kubeconfig conditions accordingly
	meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
		Type:    kubeconfigv1alpha1.ConditionTypeUserSecretFinished,
		Status:  metav1.ConditionTrue,
		Reason:  "Upserted",
		Message: "Upserted user secret with certificate from approved CSR",
	})

	crbName := types.NamespacedName{
		Name: fmt.Sprintf("%s-kubeconfig", kubeconfig.Spec.Username),
	}
	crb := &rbacv1.ClusterRoleBinding{}
	err = r.Get(ctx, crbName, crb)
	if apierrors.IsNotFound(err) {
		crb = r.createClusterRoleBinding(kubeconfig, crbName)
		err = r.Create(ctx, crb)
		if err != nil {
			klog.ErrorS(err, "failed to create clusterrolebinding object at API server, requeueing")
			return ctrl.Result{Requeue: true}, err
		}
		klog.V(2).InfoS("Created clusterrolebinding for kubeconfig user", "user", kubeconfig.Spec.Username)
	} else if err != nil {
		klog.ErrorS(err, "failed to get clusterrolebinding", "name", crbName.Name)
		r.Recorder.Eventf(kubeconfig, "Warning", "ClusterRoleBindingFailed", "Failed to get clusterrolebinding, %v", err)
		return ctrl.Result{}, err
	}

	meta.SetStatusCondition(&kubeconfig.Status.Conditions, metav1.Condition{
		Type:    kubeconfigv1alpha1.ConditionTypeKubeconfigFinished,
		Reason:  "Finished",
		Message: "Finished kubeconfig creation",
		Status:  metav1.ConditionTrue,
	})
	kubeconfig.Status.Status = phases.PhaseDone
	r.Status().Update(ctx, kubeconfig)

	klog.V(2).InfoS("Finished kubeconfig reconciliation", "name", kubeconfig.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeconfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeconfigv1alpha1.Kubeconfig{}).
		Owns(&certificatesv1.CertificateSigningRequest{}).
		Complete(r)
}
