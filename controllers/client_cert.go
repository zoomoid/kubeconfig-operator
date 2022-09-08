package controllers

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	l := log.FromContext(ctx)

	clientKeySecret := &corev1.Secret{}
	err := r.Get(ctx, object, clientKeySecret)
	if err != nil {
		return "", "", err
	}
	clientKey := clientKeySecret.Data["tls.key"]
	clientCert, ok := clientKeySecret.Data["tls.crt"]
	if !ok {
		l.V(2).Error(errors.New("no certificate found in secret"), "Secret does not contain certificate key")
	}
	return string(clientKey), string(clientCert), nil
}
