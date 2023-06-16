package v1

import "k8s.io/apimachinery/pkg/runtime"

type BindingSpec struct {
	Role     string            `json:"role"`
	Provider string            `json:"provider"`
	Config   map[string]string `json:"config,omitempty"`
}

type TopologySpec struct {
	Device   string            `json:"device,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
	Bindings []BindingSpec     `json:"bindings,omitempty"`
}

type ErrorType struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ProvisioningStatus defines the state of the ARM resource for long running operations
type ProvisioningStatus struct {
	OperationID  string            `json:"operationId"`
	Status       string            `json:"status"`
	FailureCause string            `json:"failureCause,omitempty"`
	LogErrors    bool              `json:"logErrors,omitempty"`
	Error        ErrorType         `json:"error,omitempty"`
	Output       map[string]string `json:"output,omitempty"`
}

type FilterSpec struct {
	Direction  string            `json:"direction"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type ConstraintSpec struct {
	Key       string   `json:"key"`
	Qualifier string   `json:"qualifier,omitempty"`
	Operator  string   `json:"operator,omitempty"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"` //TODO: It seems kubebuilder has difficulties handling recursive defs. This is supposed to be an ConstraintSpec array
}

type ComponentSpec struct {
	Name     string            `json:"name"`
	Type     string            `json:"type,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties   runtime.RawExtension `json:"properties,omitempty"`
	Routes       []RouteSpec          `json:"routes,omitempty"`
	Constraints  []ConstraintSpec     `json:"constraints,omitempty"`
	Dependencies []string             `json:"dependencies,omitempty"`
	Skills       []string             `json:"skills,omitempty"`
}

type RouteSpec struct {
	Route      string            `json:"route"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Filters    []FilterSpec      `json:"filters,omitempty"`
}

// TargetSpec defines the desired state of Target
type TargetSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	DisplayName   string            `json:"displayName,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
	Components    []ComponentSpec   `json:"components,omitempty"`
	Constraints   []ConstraintSpec  `json:"constraints,omitempty"`
	Topologies    []TopologySpec    `json:"topologies,omitempty"`
	ForceRedeploy bool              `json:"forceRedeploy,omitempty"`
	Scope         string            `json:"scope,omitempty"`
	Version       string            `json:"version,omitempty"`
}
