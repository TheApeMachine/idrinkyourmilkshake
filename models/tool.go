package models

type ToolType interface {
	Execute(args map[string]any) (string, error)
	Name() string
	Description() string
	Schema() any
}

type Tool struct {
	Name        string    `json:"name" jsonschema:"description=The name of the tool,required"`
	Description string    `json:"description" jsonschema:"description=The description of the tool,required"`
	Parameters  Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required    []string  `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

type Parameter struct {
	Type       string     `json:"type" jsonschema:"description=The type of the parameter,required"`
	Properties []Property `json:"properties" jsonschema:"description=The properties of the parameter,required"`
	Required   bool       `json:"required" jsonschema:"description=Whether the parameter is required,required"`
}

type Property struct {
	Type        string `json:"type" jsonschema:"description=The type of the property,required"`
	Description string `json:"description" jsonschema:"description=The description of the property,required"`
	Name        string `json:"name" jsonschema:"description=The name of the property"`
}

func NewTool(toolType ToolType) ToolType {
	return toolType
}

func NewParameter(t string, description string, properties []Property, required bool) *Parameter {
	return &Parameter{
		Type:       t,
		Properties: properties,
		Required:   required,
	}
}
