package mapping

import (
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func Fields(processors ...pipeline.MapFunc[spec.Field]) pipeline.MapFunc[spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		updatedFields := make(map[string]spec.Field)

		for name, field := range schema.Fields {
			modifiedField := field
			for _, processor := range processors {
				modifiedField = processor(modifiedField)
			}

			updatedFields[name] = modifiedField
		}

		schema.Fields = updatedFields

		return schema
	}
}
