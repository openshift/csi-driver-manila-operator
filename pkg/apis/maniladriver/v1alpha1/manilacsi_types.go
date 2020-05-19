package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManilaDriverSpec defines the desired state of ManilaDriver
type ManilaDriverSpec struct {
}

// ManilaDriverStatus defines the observed state of ManilaDriver
type ManilaDriverStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManilaDriver is the Schema for the maniladrivers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=maniladrivers,scope=Namespaced
type ManilaDriver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManilaDriverSpec   `json:"spec,omitempty"`
	Status ManilaDriverStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManilaDriverList contains a list of ManilaDriver
type ManilaDriverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManilaDriver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManilaDriver{}, &ManilaDriverList{})
}
