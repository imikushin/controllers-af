/*
Copyright 2021 Ivan Mikushin

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

// ConfigMapCountSpec defines the desired state of ConfigMapCount
type ConfigMapCountSpec struct {
	// Selector selects ConfigMaps to count
	// +optional
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

// ConfigMapCountStatus defines the observed state of ConfigMapCount
type ConfigMapCountStatus struct {
	// ConfigMaps count - selected by .spec.selector in the current namespace
	// +optional
	ConfigMaps int `json:"configMaps"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=configmapcounts,shortName=cmc
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Selector",type=string,JSONPath=.spec.selector
// +kubebuilder:printcolumn:name="ConfigMaps",type=string,JSONPath=.status.configMaps

// ConfigMapCount is the Schema for the configmapcounts API
type ConfigMapCount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigMapCountSpec `json:"spec,omitempty"`
	// +optional
	Status ConfigMapCountStatus `json:"status"`
}

// +kubebuilder:object:root=true

// ConfigMapCountList contains a list of ConfigMapCount
type ConfigMapCountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMapCount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigMapCount{}, &ConfigMapCountList{})
}
