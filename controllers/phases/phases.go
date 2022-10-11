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

package phases

const (
	// PhasePending indicates that the kubeconfig is either in its initial unreconciled (but defaulted) state, or entered a non-recoverable state
	PhasePending = "Pending"
	// PhaseAwaitingApproval indicates that the kubeconfig's CSR is pending either automatic or manual approval by a cluster admin
	PhaseAwaitingApproval = "Awaiting Approval"
	// PhaseCreatingKubeconfig indicates that the transaction to create a user's kubeconfig is in progress
	PhaseDone = "Done"
	// PhaseFailed indicates terminal failure to reconcile the kubeconfig
	PhaseFailed = "Failed"
)
