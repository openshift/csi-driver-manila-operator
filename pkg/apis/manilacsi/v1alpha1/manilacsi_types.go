package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DriverPhase string

const (
	DriverPhaseNone     DriverPhase = ""
	DriverPhaseCreating DriverPhase = "Creating"
	DriverPhaseRunning  DriverPhase = "Running"
	DriverPhaseFailed   DriverPhase = "Failed"
	DriverPhaseNoManila DriverPhase = "Manila service is not available"
)

// ManilaCSISpec defines the desired state of ManilaCSI
type ManilaCSISpec struct {
}

// ManilaCSIStatus defines the observed state of ManilaCSI
type ManilaCSIStatus struct {
	// Phase is the driver running phase
	Phase           DriverPhase `json:"phase"`
	ControllerReady bool        `json:"controllerReady"`
	NodeReady       bool        `json:"nodeReady"`
	NFSNodeReady    bool        `json:"nfsNodeReady"`

	// Version is the current operator version
	Version string `json:"version"`
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
