package controllers

import (
	"context"
	"errors"
	"fmt"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	config "github.com/zoomoid/kubeconfig-operator/pkg/kubeconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func (r *KubeconfigReconciler) ClusterEndpoint(ctx context.Context) (string, error) {
	clusterInfoCM := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Namespace: "kube-public", Name: "cluster-info"}, clusterInfoCM)

	if err != nil {
		return "", err
	}

	kubeconfig, ok := clusterInfoCM.Data["kubeconfig"]
	if !ok {
		return "", errors.New("no kubeconfig key in cluster-info config map")
	}

	cfg, err := config.Unmarshal([]byte(kubeconfig))
	if err != nil {
		return "", err
	}

	// local cluster-info kubeconfig will only contain the current cluster without a name
	// Therefore, the map will have a Cluster at key ""
	cluster, ok := cfg.Clusters[""]
	if !ok {
		return "", errors.New("no cluster entry at default location clusters.cluster.name == ''")
	}

	return cluster.Server, nil
}

func (r *KubeconfigReconciler) makeKubeconfig(ctx context.Context, kubeconfig *kubeconfigv1alpha1.Kubeconfig) error {
	// csr.Status.Certificate contains the signed certificate from the kube-api-server,
	// Overall, we need
	//   1. The key generated during the CSR generation
	//   2. The signed certificate
	//   3. The Cluster CA Certificate obtained from kube-root-ca
	// for a Kubeconfig file
	clusterCA, err := r.ClusterCA(ctx, kubeconfig.Namespace)
	if err != nil {
		// Failed to get kube root CA, fail
		return err
	}
	clientKey, clientCert, err := r.ClientData(ctx, kubeconfig.Status.Secrets.ClientTLS)
	if err != nil {
		return err
	}

	cfg := config.NewBareConfig()

	cfg.Clusters[kubeconfig.Spec.Cluster.Name] = config.Cluster{
		CertificateAuthority: clusterCA,
		Server:               kubeconfig.Spec.Cluster.Server,
	}
	cfg.Users[kubeconfig.Spec.Username] = config.User{
		ClientCertificate: clientCert,
		ClientKey:         clientKey,
	}
	contextName := fmt.Sprintf("%s@%s", kubeconfig.Spec.Username, kubeconfig.Spec.Cluster.Name)
	cfg.Contexts[contextName] = config.Context{
		Cluster:   kubeconfig.Spec.Cluster.Name,
		Namespace: "default",
		User:      kubeconfig.Spec.Username,
	}

	serializedKubeconfig := string(cfg.Marshal())

	kubeconfig.Status.Kubeconfig = serializedKubeconfig
	kubeConfigSecret := r.kubeconfigSecret(kubeconfig, serializedKubeconfig)
	err = r.Create(ctx, kubeConfigSecret)
	if err != nil {
		kubeconfig.Status.Conditions = append(kubeconfig.Status.Conditions, metav1.Condition{
			Status:  metav1.ConditionFalse,
			Type:    "Failed",
			Reason:  "KubeconfigSecretCreation",
			Message: fmt.Sprintf("Failed to create kubeconfig secret resource %s/%s, %v", kubeConfigSecret.Namespace, kubeConfigSecret.Name, err),
		})
	}

	kubeconfig.Status.Secrets.Kubeconfig = types.NamespacedName{
		Namespace: kubeConfigSecret.Namespace,
		Name:      kubeConfigSecret.Name,
	}
	return nil
}
