package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Composite is the Schema for the compositekinds API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Composite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CompositeSpec   `json:"spec,omitempty"`
	Status CompositeStatus `json:"status,omitempty"`
}

// CompositeSpec defines the desired state of Composite
type CompositeSpec struct {
	Image string `json:"image"`
}

// CompositeStatus defines the observed state of Composite.
// It should always be reconstructable from the state of the cluster and/or outside world.
type CompositeStatus struct {
	ManagedTypes   int `json:"managedTypes"`
	ManagedObjects int `json:"managedObjects"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CompositeList contains a list of Composite
type CompositeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Composite `json:"items"`
}
