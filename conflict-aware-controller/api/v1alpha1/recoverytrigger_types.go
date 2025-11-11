/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// -------------------- SPEC --------------------
type TargetObject struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
}

type RecoveryTriggerSpec struct {
	FailureType      string         `json:"failureType,omitempty"`
	WorkflowTemplate string         `json:"workflowTemplate,omitempty"`
	TargetObjects    []TargetObject `json:"targetObjects,omitempty"`
}

// ------------------- STATUS -------------------
type RecoveryTriggerStatus struct {
	State        string       `json:"state,omitempty"`
	Reason       string       `json:"reason,omitempty"`
	StartedAt    *metav1.Time `json:"startedAt,omitempty"`
	WorkflowName string       `json:"workflowName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type RecoveryTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecoveryTriggerSpec   `json:"spec,omitempty"`
	Status RecoveryTriggerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type RecoveryTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RecoveryTrigger `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RecoveryTrigger{}, &RecoveryTriggerList{})
}
