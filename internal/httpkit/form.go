package httpkit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
)

// EncodeForm encodes a record payload into application/x-www-form-urlencoded bytes.
//
// It supports string/scalar values, repeated fields via slices, and will JSON-encode
// nested objects/arrays so they can be passed as string values.
func EncodeForm(record map[string]any) ([]byte, error) {
	values := url.Values{}

	for _, key := range sortedKeys(record) {
		val := record[key]
		if val == nil {
			continue
		}

		if err := encodeFormField(values, key, val); err != nil {
			return nil, err
		}
	}

	return []byte(values.Encode()), nil
}

func sortedKeys(record map[string]any) []string {
	keys := make([]string, 0, len(record))
	for key := range record {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func encodeFormField(values url.Values, key string, val any) error {
	switch typed := val.(type) {
	case string:
		values.Set(key, typed)
	case []string:
		for _, item := range typed {
			values.Add(key, item)
		}
	case []any:
		return encodeFormSlice(values, key, typed)
	case map[string]any:
		b, err := json.Marshal(typed)
		if err != nil {
			return err
		}

		values.Set(key, string(b))
	default:
		values.Set(key, fmt.Sprint(val))
	}

	return nil
}

func encodeFormSlice(values url.Values, key string, items []any) error {
	for _, item := range items {
		if item == nil {
			continue
		}

		switch itemTyped := item.(type) {
		case map[string]any, []any:
			b, err := json.Marshal(itemTyped)
			if err != nil {
				return err
			}

			values.Add(key, string(b))
		default:
			values.Add(key, fmt.Sprint(item))
		}
	}

	return nil
}
