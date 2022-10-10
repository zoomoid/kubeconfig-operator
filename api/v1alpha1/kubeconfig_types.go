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

package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeconfigSpec defines the desired state of Kubeconfig
type KubeconfigSpec struct {
	// Username is the name associated with the future owner of the kubeconfig. The certificate is bound to this name as Common Name,
	// and the name is used for subresources as well
	Username string `json:"username,omitempty"`

	// When wanting to use an existing CSR, add a reference to the secret containing private key and csr here
	// this field is immutable after creation
	// +optional
	ExistingCSR *SecretRef `json:"existingCSR,omitempty"`

	// to not cause cascading updates to downstream CSRs and secrets,
	// this field is immutable, which is enforced by the parallel validating webhook server
	// +optional
	AutoApproveCSR bool `json:"automaticApproval,omitempty"`

	// CSR contains the parameters for generating the private key and CSR for the kube-api-server to sign
	// +optional
	CSR *CertificateSigningRequest `json:"csr,omitempty"`

	// Cluster contains information to template into the final kubeconfig, like names and endpoints
	// +optional
	Cluster *Cluster `json:"cluster"`

	// RoleRef contains the role references that the created cluster role binding links against
	// +optional
	RoleRef *rbacv1.RoleRef `json:"roleRef,omitempty"`
}

type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type CsrRef struct {
	Name string `json:"name"`
}

type Cluster struct {
	// +kubebuilder:default=kubernetes
	Name   string `json:"name"`
	Server string `json:"server"`
}

type CertificateSigningRequest struct {
	// +kubebuilder:default="SHA256WithRSA"
	SignatureAlgorithm SignatureAlgorithm                        `json:"signatureAlgorithm,omitempty"`
	AdditionalFields   CertificateSigningRequestAdditionalFields `json:"additionalFields,omitempty"`
}

// CertificateSigningRequestAdditionalFields contains the name fields of an X.509 certificate
// Excludes the ExtraNames field because we cannot properly serialize it currently
type CertificateSigningRequestAdditionalFields struct {
	// CommonName is omitted because that is the username
	// +kubebuilder:default={}
	Country []string `json:"country,omitempty"`

	// Province of the certificate requestor
	// +kubebuilder:default={}
	Province []string `json:"province,omitempty"`

	// Locality of the certificate requestor
	// +kubebuilder:default={}
	Locality []string `json:"locality,omitempty"`

	// Organization of the certificate requestor
	// +kubebuilder:default={}
	Organization []string `json:"organization,omitempty"`

	// OrganizationalUnit of the certificate requestor
	// +kubebuilder:default={}
	OrganizationalUnit []string `json:"organizationalUnit,omitempty"`

	// TODO(feat): re-add this field in another format and add a custom marshaller in the certificate signing request generation
	// ExtraNames field is excluded because this field cannot easily be serialized by controller-gen
	// ExtraNames         []pkix.AttributeTypeAndValue `json:"extraNames,omitempty"`
}

// KubeconfigStatus defines the observed state of Kubeconfig
type KubeconfigStatus struct {
	// Secrets is a struct containing references to the secrets created by the controller
	Secrets CreatedSecrets `json:"secrets,omitempty"`

	// Csr is a name reference to the CSR created by the controller
	Csr CsrRef `json:"csr,omitempty"`

	// Condititions are metav1 conditions that track the state of the kubeconfig
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"condition"`

	// Kubeconfig contains the final kubeconfig for the user as a formatted string
	// +optional
	Kubeconfig string `json:"kubeconfig,omitempty"`

	// +kubebuilder:default="Unknown"
	Phase string `json:"phase,omitempty"`
}

type CreatedSecrets struct {
	Kubeconfig SecretRef `json:"kubeconfigRef,omitempty"`
	UserSecret SecretRef `json:"userSecretRef,omitempty"`
}

type SignatureAlgorithm string

const (
	UnknownSignatureAlgorithm SignatureAlgorithm = "Unknown"

	SHA256WithRSA    SignatureAlgorithm = "SHA256WithRSA"
	SHA384WithRSA    SignatureAlgorithm = "SHA384WithRSA"
	SHA512WithRSA    SignatureAlgorithm = "SHA512WithRSA"
	ECDSAWithSHA256  SignatureAlgorithm = "ECDSAWithSHA256"
	ECDSAWithSHA384  SignatureAlgorithm = "ECDSAWithSHA384"
	ECDSAWithSHA512  SignatureAlgorithm = "ECDSAWithSHA512"
	SHA256WithRSAPSS SignatureAlgorithm = "SHA256WithRSAPSS"
	SHA384WithRSAPSS SignatureAlgorithm = "SHA384WithRSAPSS"
	SHA512WithRSAPSS SignatureAlgorithm = "SHA512WithRSAPSS"
	PureEd25519      SignatureAlgorithm = "PureEd25519"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="For",type=string,JSONPath=`.spec.username`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// Kubeconfig is the Schema for the kubeconfigs API
type Kubeconfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeconfigSpec   `json:"spec"`
	Status KubeconfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubeconfigList contains a list of Kubeconfig
type KubeconfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kubeconfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kubeconfig{}, &KubeconfigList{})
}
