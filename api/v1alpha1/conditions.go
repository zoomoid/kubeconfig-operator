package v1alpha1

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
