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

	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := record[k]
		if v == nil {
			continue
		}

		switch typed := v.(type) {
		case string:
			values.Set(k, typed)
		case []string:
			for _, item := range typed {
				values.Add(k, item)
			}
		case []any:
			for _, item := range typed {
				if item == nil {
					continue
				}
				switch itemTyped := item.(type) {
				case map[string]any, []any:
					b, err := json.Marshal(itemTyped)
					if err != nil {
						return nil, err
					}
					values.Add(k, string(b))
				default:
					values.Add(k, fmt.Sprint(item))
				}
			}
		case map[string]any:
			b, err := json.Marshal(typed)
			if err != nil {
				return nil, err
			}
			values.Set(k, string(b))
		default:
			values.Set(k, fmt.Sprint(v))
		}
	}

	return []byte(values.Encode()), nil
}

