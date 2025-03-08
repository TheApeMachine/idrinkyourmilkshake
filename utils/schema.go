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
