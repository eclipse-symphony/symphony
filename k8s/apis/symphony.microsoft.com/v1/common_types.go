package v1

import "k8s.io/apimachinery/pkg/runtime"

// Defines a component binding for a provider
type BindingSpec struct {
	Role     string            `json:"role"`
	Provider string            `json:"provider"`
	Config   map[string]string `json:"config,omitempty"`
}

// Defines the device topology for a target or instance
type TopologySpec struct {
	Device   string            `json:"device,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
	Bindings []BindingSpec     `json:"bindings,omitempty"`
}

// Defines an error in the ARM resource for long running operations
type ErrorType struct {
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Target  string        `json:"target,omitempty"`
	Details []TargetError `json:"details,omitempty"`
}

// Defines an error for symphony target
type TargetError struct {
	Code    string           `json:"code,omitempty"`
	Message string           `json:"message,omitempty"`
	Target  string           `json:"target,omitempty"`
	Details []ComponentError `json:"details,omitempty"`
}

// Defines an error for components defined in symphony
type ComponentError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Target  string `json:"target,omitempty"`
}

// Defines the state of the ARM resource for long running operations
type ProvisioningStatus struct {
	OperationID  string            `json:"operationId"`
	Status       string            `json:"status"`
	FailureCause string            `json:"failureCause,omitempty"`
	LogErrors    bool              `json:"logErrors,omitempty"`
	Error        ErrorType         `json:"error,omitempty"`
	Output       map[string]string `json:"output,omitempty"`
}

// Defines a filter for a route
type FilterSpec struct {
	Direction  string            `json:"direction"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// Defines a desired runtime component
type ComponentSpec struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties   runtime.RawExtension `json:"properties,omitempty"`
	Routes       []RouteSpec          `json:"routes,omitempty"`
	Constraints  string               `json:"constraints,omitempty"`
	Dependencies []string             `json:"dependencies,omitempty"`
	Skills       []string             `json:"skills,omitempty"`
}

// Defines the incoming and outgoing routes of the component
type RouteSpec struct {
	Route      string            `json:"route"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Filters    []FilterSpec      `json:"filters,omitempty"`
}

// Defines the desired state of Target
type TargetSpec struct {
	DisplayName   string            `json:"displayName,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
	Components    []ComponentSpec   `json:"components,omitempty"`
	Constraints   string            `json:"constraints,omitempty"`
	Topologies    []TopologySpec    `json:"topologies,omitempty"`
	ForceRedeploy bool              `json:"forceRedeploy,omitempty"`
	Scope         string            `json:"scope,omitempty"`
	// Defines the version of a particular resource
	Version    string `json:"version,omitempty"`
	Generation string `json:"generation,omitempty"`
}
