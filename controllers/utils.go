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
)

const (
	RSAKeyLength int = 4096
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

func getCertApprovalCondition(status *certificatesv1.CertificateSigningRequestStatus) (approved bool, denied bool, failed bool) {
	for _, c := range status.Conditions {
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
