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

const (
	// ConditionTypeCSRCreated is the condition type used to indicate the status of CSR creation
	ConditionTypeCSRCreated string = "CSRCreated"
	// ConditionTypeCSRApproved indicates the status of the approval of a CSR
	ConditionTypeCSRApproved string = "CSRApproved"
	// ConditionTypeUserSecretCreated indicates the creation status of the user secret object
	ConditionTypeUserSecretCreated string = "UserSecretCreated"
	// ConditionTypeUserSecretFinished indicates if the user secret has all the required data fields
	ConditionTypeUserSecretFinished string = "UserSecretFinished"
	// ConditionTypeKubeconfigFinished indicates if the kubeconfig is complete
	ConditionTypeKubeconfigFinished string = "KubeconfigFinished"
)
