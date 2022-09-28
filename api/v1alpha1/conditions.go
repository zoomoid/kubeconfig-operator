package v1alpha1

const (
	ConditionTypeCSRCreated              string = "CSRCreated"
	ConditionTypeCSRApproved             string = "CSRApproved"
	ConditionTypeUserSecretCreated       string = "UserSecretCreated"
	ConditionTypeUserSecretFinished      string = "UserSecretFinished"
	ConditionTypeKubeconfigSecretCreated string = "KubeconfigSecretCreated"
	ConditionTypeKubeconfigFinished      string = "KubeconfigFinished"
)

// // defaultConditions adds fresh conditions if their type does not yet exist
// func defaultConditions(cl []metav1.Condition) []metav1.Condition {
// 	if meta.FindStatusCondition(cl, ConditionTypeCSRCreated) == nil {
// 		cl = append(cl, metav1.Condition{
// 			Type:               ConditionTypeCSRCreated,
// 			Status:             metav1.ConditionUnknown,
// 			LastTransitionTime: metav1.Now(),
// 			Reason:             "n/a",
// 			Message:            "n/a",
// 		})
// 	}
// 	if meta.FindStatusCondition(cl, ConditionTypeCSRApproved) == nil {
// 		cl = append(cl, metav1.Condition{
// 			Type:               ConditionTypeCSRApproved,
// 			Status:             metav1.ConditionUnknown,
// 			LastTransitionTime: metav1.Now(),
// 			Reason:             "n/a",
// 			Message:            "n/a",
// 		})
// 	}
// 	if meta.FindStatusCondition(cl, ConditionTypeUserSecretCreated) == nil {
// 		cl = append(cl, metav1.Condition{
// 			Type:               ConditionTypeUserSecretCreated,
// 			Status:             metav1.ConditionUnknown,
// 			LastTransitionTime: metav1.Now(),
// 			Reason:             "n/a",
// 			Message:            "n/a",
// 		})
// 	}
// 	if meta.FindStatusCondition(cl, ConditionTypeUserSecretFinished) == nil {
// 		cl = append(cl, metav1.Condition{
// 			Type:               ConditionTypeUserSecretFinished,
// 			Status:             metav1.ConditionUnknown,
// 			LastTransitionTime: metav1.Now(),
// 			Reason:             "n/a",
// 			Message:            "n/a",
// 		})
// 	}
// 	if meta.FindStatusCondition(cl, ConditionTypeKubeconfigSecretCreated) == nil {
// 		cl = append(cl, metav1.Condition{
// 			Type:               ConditionTypeKubeconfigSecretCreated,
// 			Status:             metav1.ConditionUnknown,
// 			LastTransitionTime: metav1.Now(),
// 			Reason:             "n/a",
// 			Message:            "n/a",
// 		})
// 	}
// 	if meta.FindStatusCondition(cl, ConditionTypeKubeconfigFinished) == nil {
// 		cl = append(cl, metav1.Condition{
// 			Type:               ConditionTypeKubeconfigFinished,
// 			Status:             metav1.ConditionUnknown,
// 			LastTransitionTime: metav1.Now(),
// 			Reason:             "n/a",
// 			Message:            "n/a",
// 		})
// 	}
// 	return cl
// }
