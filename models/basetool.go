package models

// BaseTool provides a common implementation for all tools
type BaseTool struct {
	ToolName        string    `json:"name" jsonschema:"description=The name of the tool,required"`
	ToolDescription string    `json:"description" jsonschema:"description=The description of the tool,required"`
	ToolParameters  Parameter `json:"parameters" jsonschema:"description=The parameters of the tool,required"`
	Required        []string  `json:"required" jsonschema:"description=The required parameters of the tool,required"`
}

// Name returns the tool name
func (bt *BaseTool) Name() string {
	return bt.ToolName
}

// Description returns the tool description
func (bt *BaseTool) Description() string {
	return bt.ToolDescription
}