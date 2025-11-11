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

// Scope defines what kind of object the template monitors.
type Scope string

const (
	ScopeNode    Scope = "Node"
	ScopePod     Scope = "Pod"
	ScopeCluster Scope = "Cluster"
)

// QuerySpec defines one metric query.
type QuerySpec struct {
	// Name of the metric (for status report)
	Metric string `json:"metric"`
	// PromQL or API query string
	Query string `json:"query"`
}

// MLSpec defines an optional ML model serving config.
type MLSpec struct {
	// Model name or identifier in the model store
	ModelName string `json:"modelName"`
	// Optional container image for model serving
	Image string `json:"image,omitempty"`
	// Endpoint (if model already deployed)
	Endpoint string `json:"endpoint,omitempty"`
}

// DetectionTemplateSpec defines reusable config for detection agents.
type DetectionTemplateSpec struct {
	// Scope of monitoring (Pod, Node, Cluster)
	Scope Scope `json:"scope"`

	// Interval for metric collection
	Interval metav1.Duration `json:"interval"`

	// === Option A: Prometheus-based detection ===
	PrometheusAPI string      `json:"prometheusAPI,omitempty"`
	Queries       []QuerySpec `json:"queries,omitempty"`

	// === Option B: API-based detection ===
	// K8s resource to watch (e.g., core/v1/nodes, apps/v1/deployments)
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	// FieldPath inside the resource status to evaluate (dot notation)
	FieldPath string `json:"fieldPath,omitempty"`

	// Expected value (e.g., "True" for Node Ready, "Running" for Pod)
	Expected string `json:"expected,omitempty"`

	// Rule expression (optional, can combine multiple)
	Rule string `json:"rule,omitempty"`

	// Optional ML model config
	ML *MLSpec `json:"ml,omitempty"`

	// API endpoint to trigger if anomaly detected
	TriggerAPI string `json:"triggerAPI,omitempty"`
}

// DetectionTemplateStatus provides registry info.
type DetectionTemplateStatus struct {
	Valid   bool   `json:"valid,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type DetectionTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DetectionTemplateSpec   `json:"spec,omitempty"`
	Status DetectionTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type DetectionTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DetectionTemplate `json:"items"`
}
