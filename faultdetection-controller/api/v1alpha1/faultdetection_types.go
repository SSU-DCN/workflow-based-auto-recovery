/*
Copyright 2024 DCN

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

// FaultDetectionSpec defines the desired state of FaultDetection
type FaultDetectionSpec struct {
    ClusterName string            `json:"clusterName,omitempty"` // Target cluster
    Method      string            `json:"method,omitempty"`      // Detection method: rule-based or ml-based
    Metrics     []string          `json:"metrics,omitempty"`     // List of metrics to monitor
    Thresholds  map[string]string `json:"thresholds,omitempty"` // Threshold values for metrics
}

// FaultDetectionStatus defines the observed state of FaultDetection
type FaultDetectionStatus struct {
    State       string            `json:"state,omitempty"`       // Current state of detection
    LastUpdated metav1.Time `json:"lastUpdated,omitempty"` // Last updated timestamp
    Message     string      `json:"message,omitempty"`     // Status message
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FaultDetection is the Schema for the faultdetections API
type FaultDetection struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   FaultDetectionSpec   `json:"spec,omitempty"`
    Status FaultDetectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FaultDetectionList contains a list of FaultDetection
type FaultDetectionList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []FaultDetection `json:"items"`
}
