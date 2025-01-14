/*
Copyright 2022 containeroo

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

package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AccountSpecGlobalAPIKey struct {
	// Secret name containing the API key (key must be named "apiKey")
	SecretRef v1.SecretReference `json:"secretRef"`
}

// AccountSpec defines the desired state of Account
type AccountSpec struct {
	// Email of the Cloudflare account
	Email string `json:"email"`
	// Global API key of the Cloudflare account
	GlobalAPIKey AccountSpecGlobalAPIKey `json:"globalAPIKey"`
	// Interval to check account status
	// +kubebuilder:default="5m"
	// +optional
	Interval metav1.Duration `json:"interval,omitempty"`
	// List of zone names that should be managed by cloudflare-operator
	// +optional
	ManagedZones []string `json:"managedZones,omitempty"`
}

type AccountStatusZones struct {
	// Name of the zone
	// +optional
	Name string `json:"name,omitempty"`
	// ID of the zone
	// +optional
	ID string `json:"id,omitempty"`
}

// AccountStatus defines the observed state of Account
type AccountStatus struct {
	// Phase of the Account
	// +kubebuilder:validation:Enum=Active;Failed
	// +optional
	Phase string `json:"phase,omitempty"`
	// Message if the Account authentication failed
	// +optional
	Message string `json:"message,omitempty"`
	// Zones contains all the zones of the Account
	// +optional
	Zones []AccountStatusZones `json:"zones,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Account is the Schema for the accounts API
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".spec.email"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountSpec   `json:"spec,omitempty"`
	Status AccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccountList contains a list of Account
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Account{}, &AccountList{})
}
