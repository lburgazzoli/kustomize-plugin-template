package template

type Configuration struct {
	Metadata ConfigurationMeta `json:"metadata" yaml:"metadata"`
	Spec     ConfigurationSpec `json:"spec"     yaml:"spec"`
}

type ConfigurationMeta struct {
	Name string `json:"name" yaml:"name"`
}
type ConfigurationSpec struct {
	Values map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
}
