package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *KubeconfigReconciler) ClusterCA(ctx context.Context, namespace string) (string, error) {
	kubeRootCAConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "kube-root-ca.crt"}, kubeRootCAConfigMap)
	if err != nil {
		// Failed to get kube root CA, fail
		return "", err
	}
	clusterCA := kubeRootCAConfigMap.Data["ca.crt"]
	return clusterCA, nil
}

func (r *KubeconfigReconciler) ClientData(ctx context.Context, object types.NamespacedName) (string, string, error) {
	clientKeySecret := &corev1.Secret{}
	err := r.Get(ctx, object, clientKeySecret)
	if err != nil {
		return "", "", err
	}
	clientKey := clientKeySecret.Data["tls.key"]
	clientCert := clientKeySecret.Data["tls.crt"]
	return string(clientKey), string(clientCert), nil
}
