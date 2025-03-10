package models

// ToolRegistry maintains a registry of available tools
type ToolRegistry struct {
	tools []ToolType
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: []ToolType{},
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(tool ToolType) {
	r.tools = append(r.tools, tool)
}

// GetTools returns all registered tools
func (r *ToolRegistry) GetTools() []ToolType {
	return r.tools
}