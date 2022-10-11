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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	config "github.com/zoomoid/kubeconfig-operator/pkg/kubeconfig"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	CertificateSecretPrivKeyKey = "tls.key"
	CertificateSecretCSRKey     = "tls.csr"
	CertificateSecretCertKey    = "tls.crt"
	KubeconfigKey               = "kubeconfig"
)

func (r *KubeconfigReconciler) createClusterRoleBinding(kubeconfig *kubeconfigv1alpha1.Kubeconfig, obj types.NamespacedName) *rbacv1.ClusterRoleBinding {
	labels := labelsForSubresources(kubeconfig)
	clusterrolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   obj.Name,
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

// userSecret wraps the CSR bytes in a Kubernetes secret object and sets
// the kubeconfig controller as the owner for enqueueing reconciliations of the owner
// object on updates to the object
// Since the secret is immutable, the only supported update are to metadata, in which
// case the reconciliation updates
func (r *KubeconfigReconciler) createUserSecret(kubeconfig *kubeconfigv1alpha1.Kubeconfig, obj types.NamespacedName) *corev1.Secret {
	labels := labelsForSubresources(kubeconfig)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: obj.Namespace,
			Labels:    labels,
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			CertificateSecretPrivKeyKey: "",
			CertificateSecretCSRKey:     "",
			CertificateSecretCertKey:    "",
		},
	}

	controllerutil.SetControllerReference(kubeconfig, secret, r.Scheme)
	return secret
}

// createCsr creates a new certificate signing request object and sets the kubeconfig
// controller as owner.
// If the kubeconfig's AutoApproveCSR field is set to true, sets an annotation for the csr controller to auto-approve the CSR
func (r *KubeconfigReconciler) createCsr(kubeconfig *kubeconfigv1alpha1.Kubeconfig, csrBuffer *bytes.Buffer) *certificatesv1.CertificateSigningRequest {
	labels := labelsForSubresources(kubeconfig)

	annotations := map[string]string{}
	if kubeconfig.Spec.AutoApproveCSR {
		annotations[CSRAutoApproveAnnotationKey] = "true"
	}

	csr := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconfig.Name,
			// Namespace:   "kubeconfig-operator-system",
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

// createKubeconfig creates a fresh kubeconfig from several cluster resources
// csr.Status.Certificate contains the signed certificate from the kube-api-server,
// Overall, we need (a) the key generated during the CSR generation
// (b) The signed certificate, and (c) The Cluster CA Certificate obtained from kube-root-ca
// for a Kubeconfig file to be able to authenticate to a cluster
// createKubeconfig attempts to retrieve these elements from (a) the kube-root-ca.crt configmap
// and (b) the client's private key and the approved certificate from the secret that tracks the
// client data.
func (r *KubeconfigReconciler) createKubeconfig(ctx context.Context, kubeconfig *kubeconfigv1alpha1.Kubeconfig, secret *corev1.Secret) ([]byte, error) {
	clusterCA, err := r.ClusterCA(ctx)
	if err != nil {
		// Failed to get kube root CA, fail
		return nil, err
	}
	clientKey, clientCert, err := r.ClientData(ctx, secret)
	if err != nil {
		return nil, err
	}

	cfg := config.NewBareConfig()

	cfg.Clusters = map[string]config.Cluster{
		kubeconfig.Spec.Cluster.Name: {
			CertificateAuthority: base64.StdEncoding.EncodeToString([]byte(clusterCA)),
			Server:               kubeconfig.Spec.Cluster.Server,
		},
	}
	cfg.Users = map[string]config.User{
		kubeconfig.Spec.Username: {
			ClientCertificate: base64.StdEncoding.EncodeToString([]byte(clientCert)),
			ClientKey:         base64.StdEncoding.EncodeToString([]byte(clientKey)),
		},
	}
	contextName := fmt.Sprintf("%s@%s", kubeconfig.Spec.Username, kubeconfig.Spec.Cluster.Name)
	cfg.Contexts = map[string]config.Context{
		contextName: {
			Cluster:   kubeconfig.Spec.Cluster.Name,
			Namespace: "default",
			User:      kubeconfig.Spec.Username,
		},
	}
	cfg.CurrentContext = contextName
	return cfg.Marshal(), nil
}
