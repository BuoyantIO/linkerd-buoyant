package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataPlaneUpdate is a specification for a DataPlaneUpdate resource
type DataPlaneUpdate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataPlaneUpdateSpec   `json:"spec"`
	Status DataPlaneUpdateStatus `json:"status"`
}

// DataPlaneUpdateSpec is the spec for a DataPlaneUpdate resource
type DataPlaneUpdateSpec struct {
	WorkloadSelector *metav1.LabelSelector `json:"workloadSelector"`
}

type DataPlaneUpdateStatusStatus string

// These are valid status strings for a DataPlaneUpdateStatus.
const (
	DataPlaneStatusUpToDate DataPlaneUpdateStatusStatus = "UpToDate"
	DataPlaneStatusPending  DataPlaneUpdateStatusStatus = "Pending"
	DataPlaneStatusUpdating DataPlaneUpdateStatusStatus = "Updating"
	DataPlaneStatusFailed   DataPlaneUpdateStatusStatus = "Failed"
)

// DataPlaneUpdateStatus is the status for a DataPlaneUpdate resource
type DataPlaneUpdateStatus struct {
	Status                   DataPlaneUpdateStatusStatus `json:"status"`
	Desired                  int32                       `json:"desired"`
	Current                  int32                       `json:"current"`
	LastUpdateAttempt        metav1.Time                 `json:"lastUpdateAttempt"`
	LastUpdateAttemptResult  string                      `json:"lastUpdateAttemptResult"`
	LastUpdateAttemptMessage string                      `json:"lastUpdateAttemptMessage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataPlaneUpdateList is a list of DataPlaneUpdate resources
type DataPlaneUpdateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DataPlaneUpdate `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneUpdate is a specification for a ControlPlaneUpdate resource
type ControlPlaneUpdate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneUpdateSpec   `json:"spec"`
	Status ControlPlaneUpdateStatus `json:"status"`
}

// ControlPlaneUpdateSpec is the spec for a ControlPlaneUpdate resource
type ControlPlaneUpdateSpec struct {
	Components *Components `json:"components,omitempty"`
}

// Components describes the parameters desired for the control plane
// and extensions releases
type Components struct {
	Linkerd             *HelmReleaseParams `json:"linkerd"`
	LinkerdViz          *HelmReleaseParams `json:"linkerdViz"`
	LinkerdMulticluster *HelmReleaseParams `json:"linkerdMulticluster"`
	LinkerdJaeger       *HelmReleaseParams `json:"linkerdJaeger"`
	LinkerdSmi          *HelmReleaseParams `json:"linkerdSmi"`
}

// HelmReleaseParams contains the description of the target install
type HelmReleaseParams struct {
	Version string `json:"version,omitempty"`
}

type ControlPlaneUpdateStatusStatus string

// These are valid status strings for a DataPlaneUpdateStatus.
const (
	ControlPlaneStatusUpToDate ControlPlaneUpdateStatusStatus = "UpToDate"
	ControlPlaneStatusPending  ControlPlaneUpdateStatusStatus = "Pending"
	ControlPlaneStatusUpdating ControlPlaneUpdateStatusStatus = "Updating"
	ControlPlaneStatusFailed   ControlPlaneUpdateStatusStatus = "Failed"
)

// ControlPlaneUpdateStatus is the status for a ControlPlaneUpdate resource
type ControlPlaneUpdateStatus struct {
	Status                   ControlPlaneUpdateStatusStatus `json:"status"`
	Desired                  string                         `json:"desired"`
	Current                  string                         `json:"current"`
	LastUpdateAttempt        metav1.Time                    `json:"lastUpdateAttempt"`
	LastUpdateAttemptResult  string                         `json:"lastUpdateAttemptResult"`
	LastUpdateAttemptMessage string                         `json:"lastUpdateAttemptMessage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneUpdateList is a list of ControlPlaneUpdate resources
type ControlPlaneUpdateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ControlPlaneUpdate `json:"items"`
}
