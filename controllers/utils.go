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
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"

	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
	certificatesv1 "k8s.io/api/certificates/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RSAKeyLength int = 4096
)

var (
	ErrCertificateSigningRequestCreate error = apierrors.NewInternalError(errors.New("failed to create CSR"))
	ErrCertificateSigningRequestDenied error = errors.New("csr was denied")
)

func labelsForSubresources(kubeconfig *kubeconfigv1alpha1.Kubeconfig) map[string]string {
	return map[string]string{
		"kubeconfig-operator.k8s.zoomoid.dev/for":      kubeconfig.Name,
		"kubeconfig-operator.k8s.zoomoid.dev/username": kubeconfig.Spec.Username,
	}
}

func parseSignatureAlgorithm(signatureAlgorithm kubeconfigv1alpha1.SignatureAlgorithm) x509.SignatureAlgorithm {
	switch signatureAlgorithm {
	case kubeconfigv1alpha1.SHA256WithRSA:
		return x509.SHA256WithRSA
	case kubeconfigv1alpha1.SHA384WithRSA:
		return x509.SHA384WithRSA
	case kubeconfigv1alpha1.SHA512WithRSA:
		return x509.SHA512WithRSA
	case kubeconfigv1alpha1.PureEd25519:
		return x509.PureEd25519
	case kubeconfigv1alpha1.ECDSAWithSHA512:
		return x509.ECDSAWithSHA512
	case kubeconfigv1alpha1.ECDSAWithSHA384:
		return x509.ECDSAWithSHA384
	case kubeconfigv1alpha1.ECDSAWithSHA256:
		return x509.ECDSAWithSHA256
	case kubeconfigv1alpha1.SHA256WithRSAPSS:
		return x509.SHA256WithRSAPSS
	case kubeconfigv1alpha1.SHA384WithRSAPSS:
		return x509.SHA384WithRSAPSS
	case kubeconfigv1alpha1.SHA512WithRSAPSS:
		return x509.SHA512WithRSAPSS
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

func getCertApprovalCondition(conditions []certificatesv1.CertificateSigningRequestCondition) (approved bool, denied bool, failed bool) {
	for _, c := range conditions {
		if c.Type == certificatesv1.CertificateApproved {
			approved = true
		}
		if c.Type == certificatesv1.CertificateDenied {
			denied = true
		}
		if c.Type == certificatesv1.CertificateFailed {
			failed = true
		}
	}
	return
}

func createRSAKey() (*rsa.PrivateKey, *bytes.Buffer) {
	key := &bytes.Buffer{}
	priv, _ := rsa.GenerateKey(rand.Reader, RSAKeyLength)
	pem.Encode(key, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return priv, key
}

func createECDSAKey() (*ecdsa.PrivateKey, *bytes.Buffer) {
	key := &bytes.Buffer{}
	priv, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	marshalledKey, _ := x509.MarshalECPrivateKey(priv)
	pem.Encode(key, &pem.Block{Type: "EC PRIVATE KEY", Bytes: marshalledKey})
	return priv, key
}

func createEd25519Key() (ed25519.PrivateKey, *bytes.Buffer) {
	key := &bytes.Buffer{}
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	marshalledKey, _ := x509.MarshalPKCS8PrivateKey(priv)
	pem.Encode(key, &pem.Block{Type: "PRIVATE KEY", Bytes: marshalledKey})
	return priv, key
}

// createCSR creates a new PEM certificate signing request and a private key depending on what signature algorithm the kubeconfig resource specifieds
// it returns both as a buffer, or nil, and an error
func (r *KubeconfigReconciler) createCSR(kubeconfig *kubeconfigv1alpha1.Kubeconfig) (key *bytes.Buffer, csr *bytes.Buffer, err error) {
	var keyBytes crypto.Signer
	var encoded *bytes.Buffer
	csr = &bytes.Buffer{}
	sigScheme := kubeconfig.Spec.CSR.SignatureAlgorithm
	switch sigScheme {
	case kubeconfigv1alpha1.SHA256WithRSA, kubeconfigv1alpha1.SHA384WithRSA, kubeconfigv1alpha1.SHA512WithRSA, kubeconfigv1alpha1.SHA256WithRSAPSS, kubeconfigv1alpha1.SHA384WithRSAPSS, kubeconfigv1alpha1.SHA512WithRSAPSS:
		keyBytes, encoded = createRSAKey()
	case kubeconfigv1alpha1.ECDSAWithSHA256, kubeconfigv1alpha1.ECDSAWithSHA384, kubeconfigv1alpha1.ECDSAWithSHA512:
		keyBytes, encoded = createECDSAKey()
	case kubeconfigv1alpha1.PureEd25519:
		keyBytes, encoded = createEd25519Key()
	case kubeconfigv1alpha1.UnknownSignatureAlgorithm:
		return nil, nil, errors.New("unsupported signature scheme for key generation")
	}

	csrSpec := kubeconfig.Spec.CSR
	fields := csrSpec.AdditionalFields
	subj := pkix.Name{
		CommonName:         kubeconfig.Spec.Username,
		Country:            fields.Country,
		Province:           fields.Province,
		Locality:           fields.Locality,
		Organization:       fields.Organization,
		OrganizationalUnit: fields.OrganizationalUnit,
		ExtraNames:         []pkix.AttributeTypeAndValue{},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: parseSignatureAlgorithm(csrSpec.SignatureAlgorithm),
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	if err != nil {
		return nil, nil, ErrCertificateSigningRequestCreate
	}

	pem.Encode(csr, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return encoded, csr, nil
}

// isInTerminalCondition returns true if any of the relevant final conditions have reached a terminal state
// This is used to determine if a resource should be reconciled or not
func isInTerminalCondition(kubeconfig *kubeconfigv1alpha1.Kubeconfig) bool {
	// c := meta.FindStatusCondition(kubeconfig.Status.Conditions, kubeconfigv1alpha1.ConditionTypeCSRApproved)
	// u := meta.FindStatusCondition(kubeconfig.Status.Conditions, kubeconfigv1alpha1.ConditionTypeUserSecretFinished)
	f := meta.FindStatusCondition(kubeconfig.Status.Conditions, kubeconfigv1alpha1.ConditionTypeKubeconfigFinished)
	if f == nil {
		return false
	}
	return f.Status != metav1.ConditionUnknown
}
