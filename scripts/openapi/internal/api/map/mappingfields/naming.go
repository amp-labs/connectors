package mappingfields

import (
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func DisplayNameFormat(processors ...TextProcessor) pipeline.MapFunc[spec.Field] {
	return func(field spec.Field) spec.Field {
		displayName := field.DisplayName
		for _, processor := range processors {
			displayName = processor(field.URLPath, displayName)
		}

		field.DisplayName = displayName

		return field
	}
}

type TextProcessor func(urlPath, name string) string
