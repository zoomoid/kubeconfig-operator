package utils

import (
	"context"
	"errors"

	config "github.com/zoomoid/kubeconfig-operator/pkg/kubeconfig"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ClusterEndpoint(ctx context.Context, client client.Client) (string, error) {
	clusterInfoCM := &corev1.ConfigMap{}
	err := client.Get(ctx, types.NamespacedName{Namespace: "kube-public", Name: "cluster-info"}, clusterInfoCM)

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
