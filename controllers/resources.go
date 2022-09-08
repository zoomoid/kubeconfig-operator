package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *KubeconfigReconciler) clusterRoleBinding(kubeconfig *kubeconfigv1alpha1.Kubeconfig) *rbacv1.ClusterRoleBinding {
	labels := labelsForSubresources(kubeconfig)
	clusterrolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-kubeconfig", kubeconfig.Spec.Username),
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "User",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     kubeconfig.Spec.Username,
			},
		},
		RoleRef: *kubeconfig.Spec.RoleRef, // This is defaulted by the mutating webhook
	}
	return clusterrolebinding
}

// certificateSecret wraps the CSR bytes in a Kubernetes secret object and sets
// the kubeconfig controller as the owner for enqueueing reconciliations of the owner
// object on updates to the object
// Since the secret is immutable, the only supported update are to metadata, in which
// case the reconciliation updates
func (r *KubeconfigReconciler) certificateSecret(kubeconfig *kubeconfigv1alpha1.Kubeconfig, key *bytes.Buffer, csr *bytes.Buffer) *corev1.Secret {
	labels := labelsForSubresources(kubeconfig)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-client-key", kubeconfig.Name),
			Namespace: kubeconfig.Namespace,
			Labels:    labels,
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			"tls.key": base64.StdEncoding.EncodeToString(key.Bytes()),
			"tls.csr": base64.StdEncoding.EncodeToString(csr.Bytes()),
			"tls.crt": "",
		},
	}

	controllerutil.SetControllerReference(kubeconfig, secret, r.Scheme)
	return secret
}

// csr creates a new certificate signing request object and sets the kubeconfig
// controller as owner
// If the kubeconfig's AutoApproveCSR field is set to true, 1
func (r *KubeconfigReconciler) csr(kubeconfig *kubeconfigv1alpha1.Kubeconfig, csrBuffer *bytes.Buffer) *certificatesv1.CertificateSigningRequest {
	labels := labelsForSubresources(kubeconfig)

	annotations := map[string]string{}
	if kubeconfig.Spec.AutoApproveCSR {
		annotations[CSRAutoApproveAnnotationKey] = "true"
	}

	csr := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:        kubeconfig.Name,
			Namespace:   kubeconfig.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:    csrBuffer.Bytes(),
			SignerName: "kubernetes.io/kube-apiserver-client",
			Usages: []certificatesv1.KeyUsage{
				certificatesv1.UsageClientAuth,
			},
		},
	}

	controllerutil.SetControllerReference(kubeconfig, csr, r.Scheme)
	return csr
}

func (r *KubeconfigReconciler) kubeconfigSecret(kubeconfig *kubeconfigv1alpha1.Kubeconfig) *corev1.Secret {
	labels := labelsForSubresources(kubeconfig)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubeconfig", kubeconfig.Name),
			Namespace: kubeconfig.Namespace,
			Labels:    labels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}

	controllerutil.SetControllerReference(kubeconfig, secret, r.Scheme)
	return secret
}
