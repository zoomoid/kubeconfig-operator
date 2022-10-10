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
	"errors"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// ClusterCA fetches the kube-root-ca configmap from the kube-public namespace (which is statically available)
// and returns the CA certificate as a string
func (r *KubeconfigReconciler) ClusterCA(ctx context.Context) (string, error) {
	kubeRootCAConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Namespace: "kube-public", Name: "kube-root-ca.crt"}, kubeRootCAConfigMap)
	if err != nil {
		// Failed to get kube root CA, fail
		return "", err
	}
	clusterCA := kubeRootCAConfigMap.Data["ca.crt"]
	return clusterCA, nil
}

// ClientData gets the user's certificate and its private key from the user secret and returns them as string
func (r *KubeconfigReconciler) ClientData(ctx context.Context, object kubeconfigv1alpha1.SecretRef) (string, string, error) {
	clientKeySecret := &corev1.Secret{}

	err := r.Get(ctx, types.NamespacedName{
		Namespace: object.Namespace,
		Name:      object.Name,
	}, clientKeySecret)
	if err != nil {
		return "", "", err
	}
	clientKey := clientKeySecret.Data[CertificateSecretPrivKeyKey]
	clientCert, ok := clientKeySecret.Data[CertificateSecretCertKey]
	if !ok || len(clientCert) == 0 {
		klog.V(2).ErrorS(errors.New("no certificate found in secret"), "Secret does not contain certificate key")
	}
	return string(clientKey), string(clientCert), nil
}
