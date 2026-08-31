package mapping

import (
	"strings"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/iancoleman/strcase"
)

func DisplayNameFromObjectName(schema spec.Schema) spec.Schema {
	schema.DisplayName = schema.ObjectName

	return schema
}

func DisplayNameFormat(processors ...TextProcessor) pipeline.MapFunc[spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		displayName := schema.DisplayName

		for _, processor := range processors {
			displayName = processor(displayName)
		}

		schema.DisplayName = displayName

		return schema
	}
}

type TextProcessor func(name string) string

// CapitalizeFirstLetterEveryWord makes all words start with capital except some prepositions.
func CapitalizeFirstLetterEveryWord(displayName string) string {
	return naming.CapitalizeFirstLetterEveryWord(displayName)
}

// CamelCaseToSpaceSeparated converts camel case into lower case space separated string.
func CamelCaseToSpaceSeparated(displayName string) string {
	return strcase.ToDelimited(displayName, ' ')
}

// SlashesToSpaceSeparated replaces URL slashes with spaces.
func SlashesToSpaceSeparated(displayName string) string {
	return strings.ReplaceAll(displayName, "/", " ")
}
