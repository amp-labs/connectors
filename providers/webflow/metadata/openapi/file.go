package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"gopkg.in/yaml.v3"
)

var (
	// Static file containing the Webflow Data API v2 OpenAPI spec.
	// Source: https://github.com/webflow/openapi-spec (openapi/v2.yml).
	//
	//go:embed v2.yml
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](mustNormalize(apiFile)) // nolint:gochecknoglobals
)

// mustNormalize rewrites the few OpenAPI 3.1 constructs in the Webflow spec that
// kin-openapi (an OpenAPI 3.0 loader) rejects or that api3 cannot interpret.
// The embedded file is kept verbatim; only the in-memory copy is adjusted.
//
//   - The shared error schema declares a tuple `items: [{type: string}, {type: object}]`,
//     inlined into every operation's error responses. kin-openapi expects `items`
//     to be a single schema, so the tuple becomes `items: {oneOf: [...]}`.
//   - Several inlined item schemas declare `properties` without `type: object`
//     (valid in 3.1). api3 needs the explicit type to recognise an object, so it
//     is added.
//   - Some schemas carry a named `examples` map (media-type style examples placed
//     on the schema). kin-openapi expects a schema's `examples` to be a list, so
//     map-valued `examples` on schema nodes are dropped; they carry no field data.
func mustNormalize(file []byte) []byte {
	var document any
	if err := yaml.Unmarshal(file, &document); err != nil {
		panic(err)
	}

	normalizeNode(document, false)

	normalized, err := yaml.Marshal(document)
	if err != nil {
		panic(err)
	}

	return normalized
}

// normalizeNode walks the YAML document. inPropertiesMap is true when node is the
// value of a schema's `properties` key: that map's keys are field names (which may
// themselves be called "type", "items" or "properties"), not schema keywords.
func normalizeNode(node any, inPropertiesMap bool) {
	switch value := node.(type) {
	case map[string]any:
		if inPropertiesMap {
			for _, child := range value {
				normalizeNode(child, false)
			}

			return
		}

		normalizeSchemaKeywords(value)

		for key, child := range value {
			switch key {
			case "example", "examples":
				// Example payloads are data, not schema; leave them untouched.
				continue
			case "properties":
				normalizeNode(child, true)
			default:
				normalizeNode(child, false)
			}
		}
	case []any:
		for _, child := range value {
			normalizeNode(child, false)
		}
	}
}

func normalizeSchemaKeywords(value map[string]any) {
	if tuple, isTuple := value["items"].([]any); isTuple {
		value["items"] = map[string]any{"oneOf": tuple}
	}

	if _, hasProperties := value["properties"].(map[string]any); hasProperties {
		if _, hasType := value["type"]; !hasType {
			value["type"] = "object"
		}
	}

	if _, isMap := value["examples"].(map[string]any); isMap && isSchemaNode(value) {
		delete(value, "examples")
	}
}

// isSchemaNode reports whether a YAML map looks like a JSON Schema object rather
// than a media type, parameter or other OpenAPI object that legitimately holds
// a map of named examples.
func isSchemaNode(node map[string]any) bool {
	for _, key := range []string{"type", "properties", "allOf", "oneOf", "anyOf", "items", "enum"} {
		if _, ok := node[key]; ok {
			return true
		}
	}

	return false
}
