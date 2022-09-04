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
	"crypto/x509"
	"encoding/pem"
	"fmt"

	errs "errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

const (
	CSRAutoApproveAnnotationKey          string = "kubeconfig-operator.k8s.zoomoid.dev/auto-approve"
	x509TypeCerticateRequest             string = "CERTIFICATE REQUEST"
	KubeconfigOperatorAPIVersionV1Alpha1 string = "kubeconfig-operator.k8s.zoomoid.dev/v1alpha1"
)

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

	approved, denied, failed := getCertApprovalCondition(&csr.Status)
	isPending := (!approved && !denied && !failed)
	if hasAutoApproveAnnotation(csr) && isPending {
		_ = r.ApproveCSR(ctx, csr)

		err = r.Update(ctx, csr)
		// The update call to the CSRs status will trigger the next reconciliation, at which point
		// isPending will be false, thus this branch will be skipped
		return ctrl.Result{}, err
	}

	if !approved {
		// Not yet approved, may have failed or been denied or remain in intermediate state
		// Wait until the next update with reconciliation
		return ctrl.Result{}, nil
	}

	err = r.UpdateTrackingSecret(ctx, csr)

	// There's nothing else to be done for this CSR, upstream kubeconfig reconciler will do the rest
	// The CSR object will decay after 24h after approval, after which only a trace condition will remain in
	// the kubeconfig CR
	return ctrl.Result{}, err
}

func (r *CertificateSigningRequestReconciler) UpdateTrackingSecret(ctx context.Context, csr *certificatesv1.CertificateSigningRequest) error {
	l := log.FromContext(ctx)

	// CSR has been approved, extract signed client certificate
	clientCertificate := csr.Status.Certificate

	// Derive the secret name
	secretName := fmt.Sprintf("%s-%s", csr.Name, "client-key")

	clientTLSSecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: csr.Namespace}, clientTLSSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Error(err, "Could not find TLS secret for storing client-key and client-certificate after approval of CSR", "secretName", secretName, "namespace", csr.Namespace)
		}
		return err
	}

	clientTLSSecret.Data["tls.crt"] = clientCertificate

	err = r.Update(ctx, clientTLSSecret)
	return err
}

func (r *CertificateSigningRequestReconciler) ApproveCSR(ctx context.Context, csr *certificatesv1.CertificateSigningRequest) error {
	l := log.FromContext(ctx)
	_, err := parseCSR(csr.Spec.Request)

	if err != nil {
		l.V(0).Error(err, "Failed to parse x509 CSR from request field")
		csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
			Type:               certificatesv1.CertificateFailed,
			Status:             corev1.ConditionTrue,
			Reason:             "CSR failed due to malformed object",
			Message:            err.Error(),
			LastUpdateTime:     metav1.Now(),
			LastTransitionTime: metav1.Time{},
		})

		return err
	}

	csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
		Type:               certificatesv1.CertificateApproved,
		Status:             corev1.ConditionTrue,
		Reason:             "CSR was approved",
		Message:            "CSR automatically approved by csr-controller",
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Time{},
	})

	return nil
}

func (r *CertificateSigningRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certificatesv1.CertificateSigningRequest{}).
		Complete(r)
}

func hasAutoApproveAnnotation(csr *certificatesv1.CertificateSigningRequest) bool {
	v, ok := csr.Annotations[CSRAutoApproveAnnotationKey]
	return ok && v == "true"
}

func isKubeconfigCSR(csr *certificatesv1.CertificateSigningRequest) bool {
	for _, o := range csr.OwnerReferences {
		if o.Kind == "Kubeconfig" && o.APIVersion == KubeconfigOperatorAPIVersionV1Alpha1 {
			return true
		}
	}
	return false
}

func parseCSR(pemBytes []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(pemBytes)

	if block == nil || block.Type != x509TypeCerticateRequest {
		return nil, errs.New("PEM block type must be CERTIFICATE REQUEST")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)

	if err != nil {
		return nil, err
	}

	return csr, nil
}
