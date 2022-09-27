package v1alpha1

import (
	m "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// type Condition metav1.Condition

// type ConditionList []Condition

var (
	ConditionTypeCSRCreated              string = "CSRCreated"
	ConditionTypeCSRApproved             string = "CSRApproved"
	ConditionTypeUserSecretCreated       string = "UserSecretCreated"
	ConditionTypeUserSecretFinished      string = "UserSecretFinished"
	ConditionTypeKubeconfigSecretCreated string = "KubeconfigSecretCreated"
	ConditionTypeKubeconfigFinished      string = "KubeconfigFinished"
)

func (s *KubeconfigStatus) defaultConditions() {
	s.Conditions = DefaultConditions(s.Conditions)
}

// DefaultConditions adds fresh conditions if their type does not yet exist
func DefaultConditions(cl []metav1.Condition) []metav1.Condition {
	if m.FindStatusCondition(cl, ConditionTypeCSRCreated) == nil {
		cl = append(cl, metav1.Condition{
			Type:               ConditionTypeCSRCreated,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		})
	}
	if m.FindStatusCondition(cl, ConditionTypeCSRApproved) == nil {
		cl = append(cl, metav1.Condition{
			Type:               ConditionTypeCSRApproved,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		})
	}
	if m.FindStatusCondition(cl, ConditionTypeUserSecretCreated) == nil {
		cl = append(cl, metav1.Condition{
			Type:               ConditionTypeUserSecretCreated,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		})
	}
	if m.FindStatusCondition(cl, ConditionTypeUserSecretFinished) == nil {
		cl = append(cl, metav1.Condition{
			Type:               ConditionTypeUserSecretFinished,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		})
	}
	if m.FindStatusCondition(cl, ConditionTypeKubeconfigSecretCreated) == nil {
		cl = append(cl, metav1.Condition{
			Type:               ConditionTypeKubeconfigSecretCreated,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		})
	}
	if m.FindStatusCondition(cl, ConditionTypeKubeconfigFinished) == nil {
		cl = append(cl, metav1.Condition{
			Type:               ConditionTypeKubeconfigFinished,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		})
	}
	return cl
}
