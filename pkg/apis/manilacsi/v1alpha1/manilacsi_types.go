package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManilaCSISpec defines the desired state of ManilaCSI
type ManilaCSISpec struct {
}

// ManilaCSIStatus defines the observed state of ManilaCSI
type ManilaCSIStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManilaCSI is the Schema for the manilacsis API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=manilacsis,scope=Namespaced
type ManilaCSI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManilaCSISpec   `json:"spec,omitempty"`
	Status ManilaCSIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManilaCSIList contains a list of ManilaCSI
type ManilaCSIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManilaCSI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManilaCSI{}, &ManilaCSIList{})
}
