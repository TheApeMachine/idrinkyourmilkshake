package utils

import (
	"github.com/invopop/jsonschema"
)

// GenerateSchema generates a JSON schema for the given type
func GenerateSchema[T any]() any {
	reflector := jsonschema.Reflector{
		ExpandedStruct:            true,
		DoNotReference:            true,
		AllowAdditionalProperties: false,
	}

	schema := reflector.Reflect(new(T))
	return schema
}

// Add this function to the existing schema.go file

// GenerateToolSchema creates a standard schema for tools based on properties and required fields
func GenerateToolSchema(properties map[string]map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}
