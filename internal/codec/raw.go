// Package codec provides generic helpers for encoding and decoding.
package codec

import (
	"encoding/json"

	"github.com/amp-labs/connectors/internal/datautils"
)

// RawJSON is a generic wrapper that unmarshals JSON into both a typed
// value and a raw representation.
//
// It performs a *dual unmarshal* operation:
//
//  1. The entire JSON object is decoded into [Raw], a map[string]any
//     containing every key–value pair.
//
//  2. The same JSON is decoded into [Data], an instance of type T,
//     representing the known structured fields defined by the schema.
//
// This allows connectors to safely enrich or inspect
// payloads that may contain user-defined or forward-compatible fields
// without losing type safety for known properties.
//
// When marshaled back to JSON, RawJSON merges the contents of
// [Data] and [Raw] into a single flattened JSON object. Any updates to
// Data will be reflected in the output while preserving unknown fields
// originally captured in Raw.
//
// Example:
//
//	type Contact struct {
//		codec.RawJSON[UserData]
//	}
//
//	type UserData struct {
//		ID   string `json:"id"`
//		Name string `json:"name"`
//	}
type RawJSON[T any] struct {
	// Raw stores the full decoded JSON object as a key–value map.
	Raw map[string]any `json:"-"`

	// Data stores the "typed" portion of the JSON object, decoded into T.
	Data T
}

// NewRawJSON constructs a new RawJSON[T] instance from a typed value,
// initializing both the Data and Raw representations.
func NewRawJSON[T any](data T) (*RawJSON[T], error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(dataBytes, &raw); err != nil {
		return nil, err
	}

	return &RawJSON[T]{
		Raw:  raw,
		Data: data,
	}, nil
}

// UnmarshalJSON implements [json.Unmarshaler].
// It decodes the input into both the Raw map and the typed Data field.
func (r *RawJSON[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Raw); err != nil {
		return err
	}

	return json.Unmarshal(data, &r.Data)
}

// MarshalJSON implements [json.Marshaler].
// It merges the typed Data and the captured Raw values into a single
// flattened JSON object before encoding.
func (r RawJSON[T]) MarshalJSON() ([]byte, error) {
	dataBytes, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}

	var dataMap map[string]any
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return nil, err
	}

	result, err := datautils.FromMap(r.Raw).DeepCopy()
	if err != nil {
		return nil, err
	}

	result.AddMapValues(dataMap)

	return json.Marshal(result)
}
