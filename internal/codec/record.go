// Package codec provides generic helpers for encoding and decoding.
package codec

import (
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// DecoratedRecord merges a dynamic record with a structured extension,
// producing a single flattened JSON object when marshaled.
//
// It embeds a [common.Record] containing user-defined key–value pairs
// and a typed struct [T] representing schema-bound fields that must
// coexist with those user-supplied values in the API payload.
//
// When marshaled, fields from both Record and Extension are serialized
// at the same level in the resulting JSON. This allows connectors to
// enrich arbitrary record data with well-defined metadata or attributes.
//
// Example:
//
//	type MyPayloadForRecord = codec.DecoratedRecord[RecordExtension]
//
//	type RecordExtension struct {
//		ObjectName string `json:"objectName"`
//	}
//
//	record := common.Record{"name": "Bob"}
//	extension := RecordExtension{ObjectName: "users"}
//	item := codec.DecoratedRecord[RecordExtension]{Record: record, Extension: extension}
//
//	// Output:
//	// {"name": "Bob", "objectName": "users"}
type DecoratedRecord[T any] struct {
	common.Record

	Extension T
}

func (d *DecoratedRecord[T]) MarshalJSON() ([]byte, error) {
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

	// Enhance final JSON map with properties from extension.
	datautils.FromMap(jsonProperties).AddMapValues(additionalProperties)

	// Marshall combined map.
	return json.Marshal(jsonProperties)
}
