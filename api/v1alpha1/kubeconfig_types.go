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
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubeconfigSpec defines the desired state of Kubeconfig
type KubeconfigSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	Username string `json:"username,omitempty"`
	// +optional
	// this field is also immutable
	ExistingCSR types.NamespacedName `json:"existingCSR,omitempty"`
	// +optional
	// to not cause cascading updates to downstream CSRs and secrets,
	// this field is immutable, which is enforced by the parallel validating webhook server
	AutoApproveCSR bool `json:"automaticApproval"`

	// +optional
	CSR CertificateSigningRequest `json:"csr,omitempty"`

	// +optional
	Cluster Cluster `json:"cluster"`

	// +optional
	RoleRef *rbacv1.RoleRef `json:"roleRef,omitempty"`
}

type Cluster struct {
	// +kubebuilder:default="kubernetes"
	Name   string `json:"name"`
	Server string `json:"server"`
}

type CertificateSigningRequest struct {
	// +kubebuilder:default="SHA256WithRSA"
	SignatureAlgorithm SignatureAlgorithm                        `json:"signatureAlgorithm,omitempty"`
	AdditionalFields   CertificateSigningRequestAdditionalFields `json:"additionalFields,omitempty"`
}

type CertificateSigningRequestAdditionalFields struct {
	// CommonName is omitted because that is the username
	// +kubebuilder:default=[]
	Country []string `json:"country,omitempty"`

	// +kubebuilder:default=[]
	Province []string `json:"province,omitempty"`

	// +kubebuilder:default=[]
	Locality []string `json:"locality,omitempty"`

	// +kubebuilder:default=["system:masters"]
	Organization []string `json:"organization,omitempty"`

	// +kubebuilder:default=[]
	OrganizationalUnit []string `json:"organizationalUnit,omitempty"`

	// ExtraNames field is excluded because this field cannot easily be serialized by controller-gen
	// ExtraNames         []pkix.AttributeTypeAndValue `json:"extraNames,omitempty"`
}

// KubeconfigStatus defines the observed state of Kubeconfig
type KubeconfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file^
	Secrets CreatedSecrets `json:"secrets,omitempty"`

	Conditions []Condition `json:"condition,omitempty"`

	Kubeconfig string `json:"kubeconfig,omitempty"`

	// +kubebuilder:default="unknown"
	Phase string `json:"phase,omitempty"`
}

type CreatedSecrets struct {
	Kubeconfig types.NamespacedName `json:"kubeconfig,omitempty"`
	ClientTLS  types.NamespacedName `json:"clientKey,omitempty"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kubeconfig is the Schema for the kubeconfigs API
type Kubeconfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeconfigSpec   `json:"spec,omitempty"`
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
