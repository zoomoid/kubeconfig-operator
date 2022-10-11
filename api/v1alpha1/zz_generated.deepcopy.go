//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateSigningRequest) DeepCopyInto(out *CertificateSigningRequest) {
	*out = *in
	in.AdditionalFields.DeepCopyInto(&out.AdditionalFields)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateSigningRequest.
func (in *CertificateSigningRequest) DeepCopy() *CertificateSigningRequest {
	if in == nil {
		return nil
	}
	out := new(CertificateSigningRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateSigningRequestAdditionalFields) DeepCopyInto(out *CertificateSigningRequestAdditionalFields) {
	*out = *in
	if in.Country != nil {
		in, out := &in.Country, &out.Country
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Province != nil {
		in, out := &in.Province, &out.Province
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Locality != nil {
		in, out := &in.Locality, &out.Locality
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Organization != nil {
		in, out := &in.Organization, &out.Organization
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.OrganizationalUnit != nil {
		in, out := &in.OrganizationalUnit, &out.OrganizationalUnit
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateSigningRequestAdditionalFields.
func (in *CertificateSigningRequestAdditionalFields) DeepCopy() *CertificateSigningRequestAdditionalFields {
	if in == nil {
		return nil
	}
	out := new(CertificateSigningRequestAdditionalFields)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CsrRef) DeepCopyInto(out *CsrRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CsrRef.
func (in *CsrRef) DeepCopy() *CsrRef {
	if in == nil {
		return nil
	}
	out := new(CsrRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kubeconfig) DeepCopyInto(out *Kubeconfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kubeconfig.
func (in *Kubeconfig) DeepCopy() *Kubeconfig {
	if in == nil {
		return nil
	}
	out := new(Kubeconfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Kubeconfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeconfigList) DeepCopyInto(out *KubeconfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Kubeconfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeconfigList.
func (in *KubeconfigList) DeepCopy() *KubeconfigList {
	if in == nil {
		return nil
	}
	out := new(KubeconfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeconfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeconfigSpec) DeepCopyInto(out *KubeconfigSpec) {
	*out = *in
	if in.ExistingCSR != nil {
		in, out := &in.ExistingCSR, &out.ExistingCSR
		*out = new(SecretRef)
		**out = **in
	}
	if in.CSR != nil {
		in, out := &in.CSR, &out.CSR
		*out = new(CertificateSigningRequest)
		(*in).DeepCopyInto(*out)
	}
	if in.Cluster != nil {
		in, out := &in.Cluster, &out.Cluster
		*out = new(Cluster)
		**out = **in
	}
	if in.RoleRef != nil {
		in, out := &in.RoleRef, &out.RoleRef
		*out = new(v1.RoleRef)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeconfigSpec.
func (in *KubeconfigSpec) DeepCopy() *KubeconfigSpec {
	if in == nil {
		return nil
	}
	out := new(KubeconfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeconfigStatus) DeepCopyInto(out *KubeconfigStatus) {
	*out = *in
	out.UserSecret = in.UserSecret
	out.Csr = in.Csr
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeconfigStatus.
func (in *KubeconfigStatus) DeepCopy() *KubeconfigStatus {
	if in == nil {
		return nil
	}
	out := new(KubeconfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretObjectReference) DeepCopyInto(out *SecretObjectReference) {
	*out = *in
	if in.Group != nil {
		in, out := &in.Group, &out.Group
		*out = new(Group)
		**out = **in
	}
	if in.Kind != nil {
		in, out := &in.Kind, &out.Kind
		*out = new(Kind)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(Namespace)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretObjectReference.
func (in *SecretObjectReference) DeepCopy() *SecretObjectReference {
	if in == nil {
		return nil
	}
	out := new(SecretObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretRef) DeepCopyInto(out *SecretRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretRef.
func (in *SecretRef) DeepCopy() *SecretRef {
	if in == nil {
		return nil
	}
	out := new(SecretRef)
	in.DeepCopyInto(out)
	return out
}
