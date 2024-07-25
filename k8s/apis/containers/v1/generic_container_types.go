package v1

// +kubebuilder:object:generate=true
type ContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:generate=true
type ContainerSpec struct {
}
