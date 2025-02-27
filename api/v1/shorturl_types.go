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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShortURLSpec defines the desired state of ShortURL.
type ShortURLSpec struct {
	TargetURL string `json:"targetURL"`
}

// ShortURLStatus defines the observed state of ShortURL.
type ShortURLStatus struct {
	ShortPath  string `json:"shortPath,omitempty"`
	ClickCount int    `json:"clickCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ShortURL is the Schema for the shorturls API.
type ShortURL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShortURLSpec   `json:"spec,omitempty"`
	Status ShortURLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ShortURLList contains a list of ShortURL.
type ShortURLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShortURL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShortURL{}, &ShortURLList{})
}
