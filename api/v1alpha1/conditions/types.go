package v1alpha1

import (
	kubeconfigv1alpha1 "github.com/zoomoid/kubeconfig-operator/api/v1alpha1"
)

var (
	ConditionCSRCreated kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "CSRCreated",
	}
	ConditionCSRApproved kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "CSRApproved",
	}
	ConditionCSRDecayedOrRemoved kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "CSRDecayedOrRemoved",
	}
	ConditionUserSecretCreated kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "UserSecretCreated",
	}
	ConditionUserSecretFinished kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "UserSecretFinished",
	}
	ConditionKubeconfigSecretCreated kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "KubeconfigSecretCreated",
	}
	ConditionKubeconfigFinished kubeconfigv1alpha1.Condition = kubeconfigv1alpha1.Condition{
		Type: "KubeconfigFinished",
	}
)

func Defaults(existingConditions []kubeconfigv1alpha1.Condition) []kubeconfigv1alpha1.Condition {
	// TODO implement me!
	return existingConditions
}

func CSRCreated(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionCSRCreated.Type
	return condition
}

func CSRApproved(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionCSRApproved.Type
	return condition
}

func CSRDecayedOrRemoved(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionCSRDecayedOrRemoved.Type
	return condition
}

func UserSecretCreated(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionUserSecretCreated.Type
	return condition
}

func UserSecretFinished(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionUserSecretFinished.Type
	return condition
}

func KubeconfigSecretCreated(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionKubeconfigSecretCreated.Type
	return condition
}

func KubeconfigFinished(condition kubeconfigv1alpha1.Condition) kubeconfigv1alpha1.Condition {
	condition.Type = ConditionKubeconfigFinished.Type
	return condition
}
