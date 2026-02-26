package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GpuClassList contains a list of GpuClass
type GpuClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GpuClass `json:"items"`
}

// GpuClasssSpec defines the desired state of GpuClasss
type GpuClasssSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS -- desired state of cluster
	GpuOperatorMode string            `json:"gpuOperatorMode"`
	VgpuMode        string            `json:"vgpuMode"`
	AppName         string            `json:"appName"`
	Enabled         bool              `json:"enabled"`
	Labels          map[string]string `json:"labels"` // gpu node 需要的标签
	// Identifie       string `json:"identifie"`
	// ZpkUrl          string `json:"zpkUrl"`
}

// GpuClasssStatus defines the observed state of GpuClasss.
// It should always be reconstructable from the state of the cluster and/or outside world.
type GpuClasssStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS -- observed state of cluster
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// GpuClasss is the Schema for the gpuclassses API

type GpuClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GpuClasssSpec   `json:"spec,omitempty"`
	Status GpuClasssStatus `json:"status,omitempty"`
}
