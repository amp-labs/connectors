package zendeskchat

import (
	"fmt"
	"maps"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(objectName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		switch objectName {
		case chats:
			nextURL, err := jsonquery.New(node).StringOptional("next_url")
			if err != nil {
				return "", err
			}

			if nextURL != nil {
				return *nextURL, nil
			}

			return "", nil

		default:
			counts, err := jsonquery.New(node).IntegerOptional("count")
			if err != nil {
				return "", err
			}

			if counts == nil || *counts < defaultPageSize {
				return "", nil
			}

			nextURL, err := jsonquery.New(node).StringOptional("next_page")
			if err != nil {
				return "", err
			}

			if nextURL != nil {
				return *nextURL, nil
			}

			return "", nil
		}
	}
}

// keyedRecords returns records stored as values of the "data" object keyed by their identifier.
// The key is the only place the identifier appears, so it is added to each record as "id".
// Records are ordered by key, so the output is stable across calls.
//
// Response:
//
//	{"data": {"26363624373267": {"routing": {"chat_limit": 3}, "skills": []}}}
//
// Records:
//
//	[{"id": "26363624373267", "routing": {"chat_limit": 3}, "skills": []}]
func keyedRecords(node *ajson.Node) ([]map[string]any, error) {
	data, err := jsonquery.New(node).ObjectOptional("data")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return []map[string]any{}, nil
	}

	keyed, err := jsonquery.Convertor.ObjectToMap(data)
	if err != nil {
		return nil, err
	}

	keys := slices.Sorted(maps.Keys(keyed))
	records := make([]map[string]any, 0, len(keys))

	for _, key := range keys {
		value, ok := keyed[key].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected record '%s' in field 'data' to be a map, got %T: %w", key, keyed[key], common.ErrMissingExpectedValues) //nolint:lll
		}

		record := maps.Clone(value)
		record[id] = key
		records = append(records, record)
	}

	return records, nil
}
