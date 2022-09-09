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
	"fmt"

	config "github.com/zoomoid/kubeconfig-operator/pkg/kubeconfig"
	certificatesv1 "k8s.io/api/certificates/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	conditions "github.com/zoomoid/kubeconfig-operator/api/v1alpha1/conditions"
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

// +kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeconfig.k8s.zoomoid.dev,resources=kubeconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;create;update;patch;delete

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

	approvedStatus := r.CsrIsApproved(&kubeconfig.Status)

	if approvedStatus == metav1.ConditionUnknown {
		// CSR status condition updates did not include a terminal condition yet, get the CSR object from the API
		csr := &certificatesv1.CertificateSigningRequest{}
		err = r.Get(ctx, types.NamespacedName{Namespace: kubeconfig.Namespace, Name: kubeconfig.Name}, csr)

		if apierrors.IsNotFound(err) {
			// error during CSR retrival, either not yet created or decayed/deleted
			res, err := r.csrNotFound(ctx, kubeconfig, err)
			_ = r.Update(ctx, kubeconfig)
			return res, err
		} else if err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		// This might be a reconciliation caused by an update to a CSR, check if approval was added to
		// the CSR
		csrApproved, csrDenied, csrFailed := getCertApprovalCondition(&csr.Status)
		if !csrApproved && !csrDenied && !csrFailed {
			// not reached a terminal condition on the CSR yet, skip further reconciliation
			return ctrl.Result{}, nil
		}

		// any of the terminal conditions reached
		// CSR status was changed, update Kubeconfig conditions and overall status
		if !csrApproved && (csrDenied || csrFailed) {
			// CSR was denied or failed, mark this Kubeconfig resource request as being failed
			kubeconfig.Status.UpdateCondition(conditions.CSRApproved(kubeconfigv1alpha1.Condition{
				Status:             metav1.ConditionFalse,
				Reason:             "CSRDeniedOrFailed",
				Message:            "CSR for the kubeconfig was denied",
				LastTransitionTime: metav1.Now(),
			}))
			// Update kubeconfig and exit early
			// TODO check if the secret exists before updating
			err = r.Update(ctx, kubeconfig)
			return ctrl.Result{}, err
		}
	}

	// updated approval state for this reonciler run
	kubeconfig.Status.UpdateCondition(conditions.CSRApproved(kubeconfigv1alpha1.Condition{
		Status:             metav1.ConditionTrue,
		Reason:             "CSRApproval",
		Message:            "CSR for the kubeconfig was approved",
		LastTransitionTime: metav1.Now(),
	}))

	kubeConfigSecret, create, err := r.findOrCreateKubeconfig(ctx, kubeconfig)
	if err != nil {
		kubeconfig.Status.UpdateCondition(conditions.KubeconfigSecretCreated(kubeconfigv1alpha1.Condition{
			Status:             metav1.ConditionFalse,
			Reason:             "CreationError",
			Message:            fmt.Sprintf("Failed to find or create kubeconfig secret, %v", err),
			LastTransitionTime: metav1.Now(),
		}))
	}
	if create {
		err = r.Create(ctx, kubeConfigSecret)
	} else {
		err = r.Update(ctx, kubeConfigSecret)
	}
	if err != nil {
		kubeconfig.Status.UpdateCondition(conditions.KubeconfigSecretCreated(kubeconfigv1alpha1.Condition{
			Status:             metav1.ConditionFalse,
			Reason:             "TransactionError",
			Message:            fmt.Sprintf("Failed to create kubeconfig secret resource %s/%s at API server, %v", kubeConfigSecret.Namespace, kubeConfigSecret.Name, err),
			LastTransitionTime: metav1.Now(),
		}))

		// Update kubeconfig and exit early
		err = r.Update(ctx, kubeconfig)
		return ctrl.Result{}, err
	}

	if err != nil {
		l.Error(err, "failed to create kubeconfig secret object at API server, requeueing")
		return ctrl.Result{Requeue: true}, err
	}

	crb := r.clusterRoleBinding(kubeconfig)

	err = r.Create(ctx, crb)
	if err != nil {
		l.Error(err, "failed to create clusterrolebinding object at API server, requeueing")
		return ctrl.Result{Requeue: true}, err
	}

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

func (r *KubeconfigReconciler) findOrCreateKubeconfig(ctx context.Context, kubeconfig *kubeconfigv1alpha1.Kubeconfig) (*corev1.Secret, bool, error) {
	resourceName := fmt.Sprintf("%s-kubeconfig", kubeconfig.Name)
	create := false
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Namespace: kubeconfig.Namespace, Name: resourceName}, secret)
	if err != nil && !apierrors.IsNotFound(err) {
		// Any error during retrieval except Not Found is a true error case
		return nil, false, err
	}
	if apierrors.IsNotFound(err) {
		create = false
		// If secret does not exist yet, create it
		secret = r.kubeconfigSecret(kubeconfig)
	}

	cfg, err := r.createKubeconfig(ctx, kubeconfig)
	if err != nil {
		return nil, false, err
	}

	kubeconfig.Status.Kubeconfig = string(cfg)
	secret.Data["kubeconfig"] = cfg

	kubeconfig.Status.Secrets.Kubeconfig = types.NamespacedName{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	return secret, create, nil
}

// createKubeconfig creates a fresh kubeconfig from several cluster resources
// csr.Status.Certificate contains the signed certificate from the kube-api-server,
// Overall, we need (a) the key generated during the CSR generation
// (b) The signed certificate, and (c) The Cluster CA Certificate obtained from kube-root-ca
// for a Kubeconfig file to be able to authenticate to a cluster
// createKubeconfig attempts to retrieve these elements from (a) the kube-root-ca.crt configmap
// and (b) the client's private key and the approved certificate from the secret that tracks the
// client data.
// TODO(fix) this might cause a race condition if the Kubeconfig CR is reconciled before the
// CSR controller reconciled and upserted the secret
func (r *KubeconfigReconciler) createKubeconfig(ctx context.Context, kubeconfig *kubeconfigv1alpha1.Kubeconfig) ([]byte, error) {
	clusterCA, err := r.ClusterCA(ctx, kubeconfig.Namespace)
	if err != nil {
		// Failed to get kube root CA, fail
		return nil, err
	}
	clientKey, clientCert, err := r.ClientData(ctx, kubeconfig.Status.Secrets.ClientTLS)
	if err != nil {
		return nil, err
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

	return cfg.Marshal(), nil
}

func (r *KubeconfigReconciler) csrNotFound(ctx context.Context, kubeconfig *kubeconfigv1alpha1.Kubeconfig, err error) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	if apierrors.IsNotFound(err) {
		// CSR for kubeconfig does not exist yet or has decayed
		created := r.CsrIsCreated(&kubeconfig.Status)
		if created == metav1.ConditionTrue {
			// if CSR was created previously and now is unavailable, it has decayed
			// in the meantime due to the timeouts, either approved, denied, or failed
			approved := r.CsrIsApproved(&kubeconfig.Status)

			// TODO approval should have been reported correctly, and is this only case that is valid here.
			// If the other states appeared, propagate those to the conditions list and mark the condition of
			// of the CR as final

			if approved != metav1.ConditionTrue {
				err = fmt.Errorf("CSR approval for %s/%s failed terminally", kubeconfig.Namespace, kubeconfig.Name)
				l.V(0).Error(err, "CSR is gone but no approval was reported, marking kubeconfig condition as failed")

				kubeconfig.Status.UpdateCondition(conditions.CSRApproved(kubeconfigv1alpha1.Condition{
					Reason:             "CsrCreatedButNotApprovedBeforeDecay",
					Status:             metav1.ConditionFalse,
					Message:            fmt.Sprintf("CSR was created but not approved before deleted/decayed, %v", err),
					LastTransitionTime: metav1.Now(),
				}))

				return ctrl.Result{}, err
			}
		} else {
			// create a fresh CSR
			keyBuffer, csrBuffer, err := r.createCSR(kubeconfig)
			if err != nil {
				// append failure condition to Kubeconfig object
				kubeconfig.Status.UpdateCondition(conditions.CSRCreated(kubeconfigv1alpha1.Condition{
					Reason:             "CsrCreationFailed",
					Message:            fmt.Sprintf("Failed to generate private key and certificate signing request, %v", err),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
				}))

				return ctrl.Result{}, err
			}

			// Create fresh CSR and a secret keeping track of the private/public key, the CSR, and
			csr := r.csr(kubeconfig, csrBuffer)
			secret := r.certificateSecret(kubeconfig, keyBuffer, csrBuffer)

			_ = r.Create(ctx, secret)
			_ = r.Create(ctx, csr)

			kubeconfig.Status.Secrets.ClientTLS = types.NamespacedName{
				Namespace: secret.Namespace,
				Name:      secret.Name,
			}
		}

		kubeconfig.Status.UpdateCondition(conditions.CSRCreated(kubeconfigv1alpha1.Condition{
			Reason:             "CSRCreated",
			Message:            "Created CSR for kubeconfig request",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
		}))

		// TODO: add more steps here

		return ctrl.Result{}, err
	} else {
		return ctrl.Result{}, err
	}

}
