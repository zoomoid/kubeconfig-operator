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

package utils

import (
	"context"
	"errors"

	config "github.com/zoomoid/kubeconfig-operator/pkg/kubeconfig"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterEndpoint retrieves the (public) cluster endpoint from the generic kubeconfig
// embedded in the configmap kube-public/cluster-info, or fails with an error
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
