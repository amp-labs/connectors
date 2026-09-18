package mapping

import (
	"strings"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
	"github.com/iancoleman/strcase"
)

func ObjectNameFromURLPath(schema spec.Schema) spec.Schema {
	schema.ObjectName = schema.UrlPath

	return schema
}

func ObjectNameOverride(override core.TrackedMap[string, string]) pipeline.MapFunc[spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		name, ok := override.Get(schema.UrlPath)
		if ok {
			if schema.ObjectName == name {
				// The override was not useful.
				override.MarkRedundant(schema.UrlPath)
			}

			schema.ObjectName = name
		}

		return schema
	}
}

func DisplayNameFromObjectName(schema spec.Schema) spec.Schema {
	schema.DisplayName = schema.ObjectName

	return schema
}

func DisplayNameFromURLPath(schema spec.Schema) spec.Schema {
	schema.DisplayName = schema.UrlPath

	return schema
}

func DisplayNameFromLastUriPart(schema spec.Schema) spec.Schema {
	url := strings.Trim(schema.UrlPath, "/")
	parts := strings.Split(url, "/")
	schema.DisplayName = parts[len(parts)-1]

	return schema
}

func DisplayNameOverride(override core.TrackedMap[string, string]) pipeline.MapFunc[spec.Schema] {
	return func(schema spec.Schema) spec.Schema {
		name, ok := override.Get(schema.UrlPath)
		if ok {
			if schema.DisplayName == name {
				// The override was not useful.
				override.MarkRedundant(schema.UrlPath)
			}

			schema.DisplayName = name
		}

		return schema
	}
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
