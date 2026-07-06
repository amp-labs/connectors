package codec

import (
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// DecoratedRecord merges a dynamic record with a typed extension and marshals them as one flat JSON object.
//
// Important: zero values in Extension are still marshaled unless the field is tagged with `omitempty`.
// Those zero values can overwrite values already present in Record during the merge unless that is your intention.
//
// Merge rules:
//   - Keys present only in Record are preserved in the output.
//   - Fields from Extension are added to the output.
//   - Matching keys favor Extension values.
//   - Nested structs in Extension are supported.
//   - Zero-value fields in Extension still override unless omitted from JSON.
//
// Example:
//
//	type RecordExtension struct {
//	    ObjectName string `json:"objectName"`
//	}
//
//	item := codec.NewDecoratedRecord(
//		map[string]any{
//			"name": "Bob",
//		},
//		RecordExtension{
//			ObjectName: "users",
//		})
//
//	// JSON:
//	// {"name":"Bob","objectName":"users"}
type DecoratedRecord[T any] struct {
	common.Record

	Extension T
}

// NewDecoratedRecord creates a DecoratedRecord from a base record and a typed extension.
//
// The returned value marshals both parts as a single flattened JSON object.
// Fields from the extension may override values already present in the base
// record, including nested JSON object fields.
func NewDecoratedRecord[T any](base common.Record, decoration T) *DecoratedRecord[T] {
	return &DecoratedRecord[T]{
		Record:    base,
		Extension: decoration,
	}
}

// MarshalJSON merges Record and Extension into a single JSON object.
//
// The record is copied first, then extension fields are added on top.
// If both contain the same key, the extension value replaces the record value.
func (d DecoratedRecord[T]) MarshalJSON() ([]byte, error) {
	// Create a copy of records.
	jsonProperties, err := datautils.FromMap(d.Record).DeepCopy()
	if err != nil {
		return nil, err
	}

	// Marshal the extension struct.
	extBytes, err := json.Marshal(d.Extension)
	if err != nil {
		return nil, fmt.Errorf("marshal extension: %w", err)
	}

	var additionalProperties map[string]any
	if err = json.Unmarshal(extBytes, &additionalProperties); err != nil {
		return nil, fmt.Errorf("unmarshal extension: %w", err)
	}

	deepMerge(jsonProperties, additionalProperties)

	// Marshall combined map.
	return json.Marshal(jsonProperties)
}

// deepMerge merges source into destination in place.
//
// Behavior:
//   - Keys that exist only in source are added to destination.
//   - When both values are nested map[string]any values, they are merged recursively.
//   - For all other value types, the source value overrides the destination value.
//
// This is a deep merge for JSON object trees.
func deepMerge(destination, source map[string]any) {
	for key, srcValue := range source {
		dstValue, exists := destination[key]
		// Add missing keys.
		if !exists {
			destination[key] = srcValue

			continue
		}

		dstMap, dstOK := dstValue.(map[string]any)
		srcMap, srcOK := srcValue.(map[string]any)

		// Nested maps are merged together instead of one map overriding the other.
		if dstOK && srcOK {
			deepMerge(dstMap, srcMap)
			destination[key] = dstMap

			continue
		}

		// Override value.
		destination[key] = srcValue
	}
}
