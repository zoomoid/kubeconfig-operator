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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
)

// CertificateSigningRequestReconciler reconciles a CSR object
type CertificateSigningRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;watch;list
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,resourceNames="kubernetes.io/kubelet-serving",verbs=approve

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kubeconfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *CertificateSigningRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// TODO(user): your logic here

	csr := &certificatesv1.CertificateSigningRequest{}
	err := r.Get(ctx, req.NamespacedName, csr)
	if err != nil {
		if errors.IsNotFound(err) {
			// The CSR was deleted, either by hand or by decay
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !isKubeconfigCSR(csr) {
		l.V(5).Info("non-kubeconfig CSR, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	approved, _, _ := getCertApprovalCondition(&csr.Status)
	if !approved {
		// Not yet approved, may have failed or been denied or remain in intermediate state
		// Wait until the next update with reconciliation
		return ctrl.Result{}, nil
	}

	// CSR has been approved
	clientCertificate := csr.Status.Certificate

	name := csr.Labels["kubeconfig-operator.k8s.zoomoid.dev/for"]
	clientTLSSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: name, Namespace: req.Namespace}, clientTLSSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Error(err, "Could not find TLS secret for storing client-key and client-certificate after approval of CSR", "secretName", name, "namespace", req.Namespace)
		}
		return ctrl.Result{}, err
	}

	clientTLSSecret.Data["tls.crt"] = clientCertificate

	err = r.Update(ctx, clientTLSSecret)

	return ctrl.Result{}, err
}

func (r *CertificateSigningRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certificatesv1.CertificateSigningRequest{}).
		Complete(r)
}

func isKubeconfigCSR(csr *certificatesv1.CertificateSigningRequest) bool {
	for _, o := range csr.OwnerReferences {
		if o.Kind == "Kubeconfig" && o.APIVersion == "kubeconfig-operator.k8s.zoomoid.dev/v1alpha1" {
			return true
		}
	}
	return false
}
