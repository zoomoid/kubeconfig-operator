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
	"time"

	errs "errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
)

// CertificateSigningRequestReconciler reconciles a CSR object
type CertificateSigningRequestReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	ClientSet *clientset.Clientset
}

const (
	CSRAutoApproveAnnotationKey          string = "kubeconfig.k8s.zoomoid.dev/auto-approve"
	x509TypeCerticateRequest             string = "CERTIFICATE REQUEST"
	KubeconfigOperatorAPIVersionV1Alpha1 string = "kubeconfig.k8s.zoomoid.dev/v1alpha1"
)

//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;watch;list
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,resourceNames="kubernetes.io/kube-apiserver-client",verbs=approve

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

	csr := &certificatesv1.CertificateSigningRequest{}
	err := r.Get(ctx, req.NamespacedName, csr)
	if err != nil {
		if errors.IsNotFound(err) {
			// The CSR was deleted, either by hand or by decay
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	klog.InfoS("Reconciling CSR", "name", req.Name)

	if !isKubeconfigV1Alpha1CSR(csr) {
		klog.V(0).Info("non-kubeconfig CSR, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	approved, denied, failed := getCertApprovalCondition(csr.Status.Conditions)
	isPending := (!approved && !denied && !failed)
	klog.V(5).InfoS("parsed CSR status", "name", req.Name, "approved", approved, "denied", denied, "failed", failed)

	if hasAutoApproveAnnotation(csr) && isPending {
		err = r.ApproveCSR(ctx, csr)
		if err != nil {
			r.Recorder.Eventf(csr, "Warning", "Failed", "Failed to approve CSR, %v", err)
			_ = r.Status().Update(ctx, csr)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !approved {
		// Not yet approved, may have failed or been denied or remain in intermediate state
		// Wait until the next update with reconciliation
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *CertificateSigningRequestReconciler) ApproveCSR(ctx context.Context, csr *certificatesv1.CertificateSigningRequest) error {
	_, err := parseCSR(csr.Spec.Request)
	if err != nil {
		klog.V(0).ErrorS(err, "Failed to parse x509 CSR from request field")
		setStatusCondition(&csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
			Type:    certificatesv1.CertificateFailed,
			Status:  corev1.ConditionTrue,
			Reason:  "Failed",
			Message: err.Error(),
		})
		return err
	}

	setStatusCondition(&csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
		Type:    certificatesv1.CertificateApproved,
		Status:  corev1.ConditionTrue,
		Reason:  "KubeconfigControllerApprove",
		Message: "The CSR was auto-approved by the kubeconfig operator",
	})

	csr, err = r.ClientSet.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, csr.Name, csr, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	klog.InfoS("Approved CSR", "name", csr.Name)
	return nil
}

func (r *CertificateSigningRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certificatesv1.CertificateSigningRequest{}).
		Complete(r)
}

// hasAutoApproveAnnotation checks for an annotation that marks the CSR as auto-approvable by
// the controller. If it is not set, the approval needs to be performed manually
func hasAutoApproveAnnotation(csr *certificatesv1.CertificateSigningRequest) bool {
	v, ok := csr.Annotations[CSRAutoApproveAnnotationKey]
	if ok && v == "true" {
		klog.V(5).InfoS("CSR has auto-approve annotation", "name", csr.Name)
		return true
	}
	klog.V(5).InfoS("CSR does not have auto-approve annotation", "name", csr.Name)
	return false
}

// isKubeconfigV1Alpha1CSR checks if the owner of the csr is a kubeconfig.kubeconfig-operator.k8s.zoomoid.dev/v1alpha1
func isKubeconfigV1Alpha1CSR(csr *certificatesv1.CertificateSigningRequest) bool {
	for _, o := range csr.OwnerReferences {
		if o.Kind == "Kubeconfig" && o.APIVersion == KubeconfigOperatorAPIVersionV1Alpha1 {
			return true
		}
	}
	return false
}

// parseCSR unwraps the base64 encoded PEM block
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

func setStatusCondition(conditions *[]certificatesv1.CertificateSigningRequestCondition, newCondition certificatesv1.CertificateSigningRequestCondition) {
	if conditions == nil {
		return
	}
	existingCondition := findStatusCondition(*conditions, newCondition.Type)
	if existingCondition == nil {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.NewTime(time.Now())
		}
		*conditions = append(*conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		if !newCondition.LastTransitionTime.IsZero() {
			existingCondition.LastTransitionTime = newCondition.LastTransitionTime
		} else {
			existingCondition.LastTransitionTime = metav1.NewTime(time.Now())
		}
	}
	existingCondition.Reason = newCondition.Reason
	existingCondition.Message = newCondition.Message
}

// FindStatusCondition finds the conditionType in conditions.
func findStatusCondition(conditions []certificatesv1.CertificateSigningRequestCondition, conditionType certificatesv1.RequestConditionType) *certificatesv1.CertificateSigningRequestCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}

	return nil
}
