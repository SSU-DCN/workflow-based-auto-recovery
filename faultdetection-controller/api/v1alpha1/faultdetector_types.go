/*
Copyright 2024.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FaultDetectorSpec defines the desired state of FaultDetector
type FaultDetectorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of FaultDetector. Edit faultdetector_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	TargetResource string `json:"targetResource"` // Resource to monitor
    FaultCondition string `json:"faultCondition"` // Condition for detecting a fault
    RecoveryAction string `json:"recoveryAction"` // Recovery workflow to trigger
}

// FaultDetectorStatus defines the observed state of FaultDetector
type FaultDetectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase   string `json:"phase,omitempty"`   // Current phase: Normal, FaultDetected, etc.
    Message string `json:"message,omitempty"` // Status message
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FaultDetector is the Schema for the faultdetectors API
type FaultDetector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FaultDetectorSpec   `json:"spec,omitempty"`
	Status FaultDetectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FaultDetectorList contains a list of FaultDetector
type FaultDetectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FaultDetector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FaultDetector{}, &FaultDetectorList{})
}
