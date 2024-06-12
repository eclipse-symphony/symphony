package testhelpers

// TODO: Switch over to symphony core types from the /k8s/api folder
type (
	Metadata struct {
		Annotations map[string]string `yaml:"annotations,omitempty"`
		Name        string            `yaml:"name,omitempty"`
		Namespace   string            `yaml:"namespace,omitempty"`
	}

	// Solution describes the structure of symphony solution yaml file
	Solution struct {
		ApiVersion string       `yaml:"apiVersion"`
		Kind       string       `yaml:"kind"`
		Metadata   Metadata     `yaml:"metadata"`
		Spec       SolutionSpec `yaml:"spec"`
	}

	SolutionSpec struct {
		DisplayName  string            `yaml:"displayName,omitempty"`
		Scope        string            `yaml:"scope,omitempty"`
		Metadata     map[string]string `yaml:"metadata,omitempty"`
		Components   []ComponentSpec   `yaml:"components,omitempty"`
		RootResource string            `yaml:"rootResource"`
	}

	// Target describes the structure of symphony target yaml file
	Target struct {
		ApiVersion string     `yaml:"apiVersion"`
		Kind       string     `yaml:"kind"`
		Metadata   Metadata   `yaml:"metadata"`
		Spec       TargetSpec `yaml:"spec"`
	}

	TargetSpec struct {
		DisplayName  string            `yaml:"displayName"`
		Scope        string            `yaml:"scope"`
		Components   []ComponentSpec   `yaml:"components,omitempty"`
		Topologies   []Topology        `yaml:"topologies"`
		Properties   map[string]string `yaml:"properties,omitempty"`
		RootResource string            `yaml:"rootResource"`
	}

	Topology struct {
		Bindings []Binding `yaml:"bindings"`
	}

	Binding struct {
		Config   Config `yaml:"config"`
		Provider string `yaml:"provider"`
		Role     string `yaml:"role"`
	}

	Config struct {
		InCluster string `yaml:"inCluster"`
	}

	ComponentSpec struct {
		Name         string                         `yaml:"name"`
		Parameters   map[string]ParameterDefinition `yaml:"parameters,omitempty"`
		Properties   map[string]interface{}         `yaml:"properties"`
		Type         string                         `yaml:"type"`
		Constraints  string                         `yaml:"constraints,omitempty"`
		Dependencies []string                       `yaml:"dependencies,omitempty"`
	}

	Instance struct {
		ApiVersion string       `yaml:"apiVersion"`
		Kind       string       `yaml:"kind"`
		Metadata   Metadata     `yaml:"metadata"`
		Spec       InstanceSpec `yaml:"spec"`
	}

	InstanceSpec struct {
		DisplayName  string                 `yaml:"displayName"`
		Target       TargetSelector         `yaml:"target"`
		Solution     string                 `yaml:"solution"`
		Scope        string                 `yaml:"scope"`
		Parameters   map[string]interface{} `yaml:"parameters,omitempty"`
		RootResource string                 `yaml:"rootResource"`
	}

	TargetSelector struct {
		Name     string            `yaml:"name,omitempty"`
		Selector map[string]string `yaml:"selector,omitempty"`
	}

	ParameterDefinition struct {
		Type         string      `yaml:"type"`
		DefaultValue interface{} `yaml:"default"`
	}

	InstanceContainer struct {
		ApiVersion string                `yaml:"apiVersion"`
		Kind       string                `yaml:"kind"`
		Metadata   Metadata              `yaml:"metadata"`
		Spec       InstanceContainerSpec `yaml:"spec"`
	}

	InstanceContainerSpec struct {
	}

	SolutionContainer struct {
		ApiVersion string                `yaml:"apiVersion"`
		Kind       string                `yaml:"kind"`
		Metadata   Metadata              `yaml:"metadata"`
		Spec       SolutionContainerSpec `yaml:"spec"`
	}

	SolutionContainerSpec struct {
	}

	TargetContainer struct {
		ApiVersion string              `yaml:"apiVersion"`
		Kind       string              `yaml:"kind"`
		Metadata   Metadata            `yaml:"metadata"`
		Spec       TargetContainerSpec `yaml:"spec"`
	}

	TargetContainerSpec struct {
	}
)
