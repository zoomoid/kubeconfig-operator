package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Condition metav1.Condition

type ConditionList []Condition

var (
	ConditionCSRCreated Condition = Condition{
		Type: "CSRCreated",
	}
	ConditionCSRApproved Condition = Condition{
		Type: "CSRApproved",
	}
	ConditionUserSecretCreated Condition = Condition{
		Type: "UserSecretCreated",
	}
	ConditionUserSecretFinished Condition = Condition{
		Type: "UserSecretFinished",
	}
	ConditionKubeconfigSecretCreated Condition = Condition{
		Type: "KubeconfigSecretCreated",
	}
	ConditionKubeconfigFinished Condition = Condition{
		Type: "KubeconfigFinished",
	}
)

// type Condition conditions.Condition

// UpdateCondition replaces an existing condition of the same type
// or adds it if it does not exist in the condition list yet
func (s *KubeconfigStatus) UpdateCondition(condition Condition) {
	found, index := s.conditionExists(condition)
	// If not found, append new condition
	if !found {
		s.Conditions = append(s.Conditions, condition)
		return
	}
	// Replace existing condition with new one
	s.Conditions[index] = condition
}

func (s *KubeconfigStatus) conditionExists(condition Condition) (bool, int) {
	// find existing condition type if it exists
	index := -1
	found := false
	for i, c := range s.Conditions {
		if condition.Type == c.Type {
			index = i
			found = true
			break
		}
	}
	return found, index
}

func (s *KubeconfigStatus) defaultConditions() {
	s.Conditions = Default(s.Conditions)
}

func (r ConditionList) Contains(c Condition) bool {
	for _, cl := range r {
		if c.Type == cl.Type {
			return true
		}
	}
	return false
}

// Default adds fresh conditions if their type does not yet exist
func Default(existingConditions []Condition) []Condition {
	cl := ConditionList(existingConditions)
	if !cl.Contains(ConditionCSRCreated) {
		cl = append(cl, CSRCreated(Condition{
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		}))
	}
	if !cl.Contains(ConditionCSRApproved) {
		cl = append(cl, CSRApproved(Condition{
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		}))
	}
	if !cl.Contains(ConditionUserSecretCreated) {
		cl = append(cl, UserSecretCreated(Condition{
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		}))
	}
	if !cl.Contains(ConditionUserSecretFinished) {
		cl = append(cl, UserSecretFinished(Condition{
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		}))
	}
	if !cl.Contains(ConditionKubeconfigSecretCreated) {
		cl = append(cl, KubeconfigSecretCreated(Condition{
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		}))
	}
	if !cl.Contains(ConditionKubeconfigFinished) {
		cl = append(cl, KubeconfigFinished(Condition{
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "n/a",
			Message:            "n/a",
		}))
	}
	return cl
}

func CSRCreated(condition Condition) Condition {
	condition.Type = ConditionCSRCreated.Type
	return condition
}

func CSRApproved(condition Condition) Condition {
	condition.Type = ConditionCSRApproved.Type
	return condition
}

func UserSecretCreated(condition Condition) Condition {
	condition.Type = ConditionUserSecretCreated.Type
	return condition
}

func UserSecretFinished(condition Condition) Condition {
	condition.Type = ConditionUserSecretFinished.Type
	return condition
}

func KubeconfigSecretCreated(condition Condition) Condition {
	condition.Type = ConditionKubeconfigSecretCreated.Type
	return condition
}

func KubeconfigFinished(condition Condition) Condition {
	condition.Type = ConditionKubeconfigFinished.Type
	return condition
}
