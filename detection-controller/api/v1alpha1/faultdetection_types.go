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

// FaultDetectionSpec references a template and target object.
type FaultDetectionSpec struct {
	// Which template to use
	TemplateRef string `json:"templateRef"`
	// Target object (optional, based on Scope)
	Target *ObjectRef `json:"target,omitempty"`
}

// ObjectRef describes the object being monitored
type ObjectRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
}

// FaultDetectionStatus captures monitoring results.
type NodeResult struct {
	NodeName string `json:"nodeName,omitempty"`
	Ok       bool   `json:"ok,omitempty"`
	Message  string `json:"message,omitempty"`
}

type FaultDetectionStatus struct {
	LastRun     *metav1.Time `json:"lastRun,omitempty"`
	Results     []Result     `json:"results,omitempty"`
	NodeResults []NodeResult `json:"nodeResults,omitempty"`
	Anomalous   bool         `json:"anomalous,omitempty"`
	Reason      string       `json:"reason,omitempty"`
	Triggered   bool         `json:"triggered,omitempty"`
	TriggerMsg  string       `json:"triggerMsg,omitempty"`
}

// Result stores metric query output
type Result struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type FaultDetection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FaultDetectionSpec   `json:"spec,omitempty"`
	Status FaultDetectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type FaultDetectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FaultDetection `json:"items"`
}
