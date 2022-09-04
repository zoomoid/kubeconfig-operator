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
	"errors"
	"fmt"

	certificatesv1 "k8s.io/api/certificates/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeconfigReconciler reconciles a Kubeconfig object
type KubeconfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	CertificateSigningRequestCreateError error = apierrors.NewInternalError(errors.New("failed to create CSR"))
)

//+kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kubeconfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KubeconfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	kubeconfig := &kubeconfigv1alpha1.Kubeconfig{}
	err := r.Get(ctx, req.NamespacedName, kubeconfig)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// this is a delete request, kubernetes will garbage collect all other objects
			// with ownership from this controller
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// check .Spec.Cluster.Server is not set and otherwise default it
	if kubeconfig.Spec.Cluster.Server == "" {
		server, err := r.ClusterEndpoint(ctx)
		if err != nil {
			l.Error(err, "failed to obtain cluster endpoint from cluster-info configmap, defaulting to localhost")
			server = "https://localhost:6443"
		}
		kubeconfig.Spec.Cluster.Server = server
	}

	csr := &certificatesv1.CertificateSigningRequest{}
	err = r.Get(ctx, types.NamespacedName{Namespace: kubeconfig.Namespace, Name: kubeconfig.Name}, csr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// CSR for kubeconfig does not exist yet or has decayed
			created := r.CsrIsCreated(&kubeconfig.Status)
			if created {
				// if CSR was created previously and now is unavailable, it has decayed
				// in the meantime due to the timeouts, either approved, denied, or failed

			} else {
				// create a fresh CSR
				keyBuffer, csrBuffer, err := r.createCSR(kubeconfig)
				if err != nil {
					// append failure condition to Kubeconfig object
					kubeconfig.Status.Conditions = append(kubeconfig.Status.Conditions, metav1.Condition{
						Type:    "Failed",
						Reason:  "CSRCreationFailure",
						Message: fmt.Sprintf("Failed to generate private key and certificate signing request, %v", err),
						Status:  metav1.ConditionUnknown,
					})
					return ctrl.Result{}, err
				}

				// Create fresh CSR and a secret keeping track of the private/public key, the CSR, and
				csr = r.csrForKubeconfig(kubeconfig, csrBuffer)
				secret := r.csrSecretForKubeconfig(kubeconfig, keyBuffer, csrBuffer)

				_ = r.Create(ctx, secret)
				_ = r.Create(ctx, csr)

				kubeconfig.Status.Secrets.ClientTLS = types.NamespacedName{
					Namespace: secret.Namespace,
					Name:      secret.Name,
				}
				// TODO: add more steps here

				kubeconfig.Status.Conditions = append(kubeconfig.Status.Conditions, metav1.Condition{
					Type:    string(kubeconfigv1alpha1.CertificateSigningRequestCreated),
					Reason:  "CSRCreated",
					Message: "Created CSR for kubeconfig request",
					Status:  metav1.ConditionFalse,
				})
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	// This might be a reconciliation caused by an update to a CSR, check if approval was added to
	// the CSR
	approved, denied, failed := getCertApprovalCondition(&csr.Status)
	if !approved && !denied && !failed { // not reached a terminal condition on the CSR yet, skip further reconciliation
		return ctrl.Result{}, nil
	}

	// any of the terminal conditions reached
	// CSR status was changed, update Kubeconfig conditions and overall status
	if approved {

		err = r.makeKubeconfig(ctx, kubeconfig)
		l.Error(err, "failed to create kubeconfig")
	} else {
		// CSR was denied, mark this Kubeconfig resource request as being failed
		kubeconfig.Status.Conditions = append(kubeconfig.Status.Conditions, metav1.Condition{
			Status:  metav1.ConditionFalse,
			Type:    "Failed",
			Reason:  "CSRDenial",
			Message: "",
		})
	}

	// CSR is still available and in-flight, check for its status and the current requests updated
	// any fields of the Spec, since the CSR is an immutable secret, we need to create a new CSR from
	// the updated fields

	// check that the .status.certificate field in the kubeconfig
	// equals the .status.certificate field in the CSR

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeconfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeconfigv1alpha1.Kubeconfig{}).
		Owns(&certificatesv1.CertificateSigningRequest{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// csrSecretForKubeconfig wraps the CSR bytes in a Kubernetes secret object and sets
// the kubeconfig controller as the owner for enqueueing reconciliations of the owner
// object on updates to the object
// Since the secret is immutable, the only supported update are to metadata, in which
// case the reconciliation updates
func (r *KubeconfigReconciler) csrSecretForKubeconfig(kubeconfig *kubeconfigv1alpha1.Kubeconfig, key *bytes.Buffer, csr *bytes.Buffer) *corev1.Secret {
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

func (r *KubeconfigReconciler) kubeconfigSecret(kubeconfig *kubeconfigv1alpha1.Kubeconfig, cfg string) *corev1.Secret {
	labels := labelsForSubresources(kubeconfig)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubeconfig", kubeconfig.Name),
			Namespace: kubeconfig.Namespace,
			Labels:    labels,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"kubeconfig": cfg,
		},
	}

	controllerutil.SetControllerReference(kubeconfig, secret, r.Scheme)
	return secret
}

func (r *KubeconfigReconciler) csrForKubeconfig(kubeconfig *kubeconfigv1alpha1.Kubeconfig, csrBuffer *bytes.Buffer) *certificatesv1.CertificateSigningRequest {
	labels := labelsForSubresources(kubeconfig)

	annotations := map[string]string{}
	if kubeconfig.Spec.AutoApproveCSR {
		annotations[CSRAutoApproveAnnotationKey] = "true"
	}

	csr := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeconfig.Name,
			// Namespace: kubeconfig.Namespace,
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

func labelsForSubresources(kubeconfig *kubeconfigv1alpha1.Kubeconfig) map[string]string {
	return map[string]string{
		"kubeconfig-operator.k8s.zoomoid.dev/for":      kubeconfig.Name,
		"kubeconfig-operator.k8s.zoomoid.dev/username": kubeconfig.Spec.Username,
	}
}
