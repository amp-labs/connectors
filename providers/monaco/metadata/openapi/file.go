package openapi

import (
	_ "embed"
	"encoding/json"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed openapi.json
	apiFile []byte

	FileManager = api3.NewOpenapiFileManager[any](collapseNullableAnyOf(apiFile)) // nolint:gochecknoglobals
)

// collapseNullableAnyOf rewrites `anyOf: [{...}, {"type": "null"}]` into the
// non-null member, hoisted into the parent object.
//
// Monaco marks almost every response field optional this way. A schema whose
// type lives inside an anyOf has no `type` of its own, so the extractor cannot
// classify it and falls back to valueType "other" -- which would leave the
// majority of fields untyped (e.g. contacts.email, an obvious string). Nullability
// carries no information for us: ListObjectMetadata reports what a field holds,
// not whether it may be absent.
//
// Only the unambiguous single-non-null case is collapsed. A genuine union of two
// or more real types is left intact, since "other" is the honest answer there.
// The vendored openapi.json stays untouched so it can be re-downloaded verbatim.
func collapseNullableAnyOf(spec []byte) []byte {
	var root any
	if err := json.Unmarshal(spec, &root); err != nil {
		// Leave the spec alone; the explorer will report the parse failure.
		return spec
	}

	rewritten, err := json.Marshal(walkNullableAnyOf(root))
	if err != nil {
		return spec
	}

	return rewritten
}

func walkNullableAnyOf(node any) any {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			value[key] = walkNullableAnyOf(child)
		}

		if survivor, ok := soleNonNullMember(value["anyOf"]); ok {
			delete(value, "anyOf")

			// Keys already on the parent (title, description, examples) win;
			// the survivor only contributes what the parent is missing, which
			// is the type information the extractor needs.
			for key, child := range survivor {
				if _, exists := value[key]; !exists {
					value[key] = child
				}
			}
		}

		return value
	case []any:
		for i, child := range value {
			value[i] = walkNullableAnyOf(child)
		}

		return value
	default:
		return node
	}
}

// soleNonNullMember returns the single member of an anyOf list that is not the
// null type, provided exactly one such member exists.
func soleNonNullMember(anyOf any) (map[string]any, bool) {
	members, ok := anyOf.([]any)
	if !ok {
		return nil, false
	}

	var survivors []map[string]any

	for _, member := range members {
		schema, isObject := member.(map[string]any)
		if !isObject {
			return nil, false
		}

		if schema["type"] == "null" {
			continue
		}

		survivors = append(survivors, schema)
	}

	if len(survivors) != 1 || len(survivors) == len(members) {
		return nil, false
	}

	return survivors[0], true
}
