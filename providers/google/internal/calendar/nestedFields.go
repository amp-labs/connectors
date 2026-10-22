package calendar

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Google Calendar events keep their timing, origin and tracing data inside nested objects.
// The metadata schema exposes those leaves in JSONPath notation, e.g. "$['start']['dateTime']",
// following the convention used by Gmail. Bracket notation keeps keys that contain dots
// (allowed in extendedProperties, e.g. "com.acme.dealId") as a single segment.
// https://developers.google.com/workspace/calendar/api/v3/reference/events#resource

var (
	errConflictingFields     = errors.New("conflicting fields in record data")
	errUnsupportedFieldPath  = errors.New("unsupported field path in record data")
	errFieldPathNotSupported = errors.New("array wildcards are not supported when writing")
)

// jsonPathField builds the JSONPath field name for a nested path, e.g. ["start", "dateTime"]
// becomes "$['start']['dateTime']".
func jsonPathField(segments ...string) string {
	var builder strings.Builder

	builder.WriteString("$")

	for _, segment := range segments {
		builder.WriteString("['")
		builder.WriteString(segment)
		builder.WriteString("']")
	}

	return builder.String()
}

// nestRecordData converts JSONPath keys in record data into nested objects, so that
// {"$['start']['dateTime']": "..."} becomes {"start": {"dateTime": "..."}}.
//
// Plain keys are copied first, then JSONPath keys are applied in sorted order so the result
// is deterministic. A JSONPath key merges into an existing nested object. It is an error
// when two keys overlap, e.g. "start" holding a string and "$['start']['dateTime']" both
// present, or when the same leaf is set twice.
func nestRecordData(data map[string]any) (map[string]any, error) {
	result := make(map[string]any, len(data))

	keys := slices.Sorted(func(yield func(string) bool) {
		for key := range data {
			if !yield(key) {
				return
			}
		}
	})

	var nested []string

	for _, key := range keys {
		path := readhelper.ParseJSONPath(key)
		if len(path) == 1 {
			result[key] = data[key]
		} else {
			nested = append(nested, key)
		}
	}

	for _, key := range nested {
		if err := setNestedValue(result, readhelper.ParseJSONPath(key), data[key]); err != nil {
			return nil, fmt.Errorf("%w: %v", err, key)
		}
	}

	return result, nil
}

func setNestedValue(target map[string]any, path []string, value any) error {
	if slices.Contains(path, readhelper.ArrayWildIndex) {
		return errFieldPathNotSupported
	}

	if len(path) == 0 {
		return errUnsupportedFieldPath
	}

	current := target

	for _, segment := range path[:len(path)-1] {
		existing, found := current[segment]
		if !found {
			next := make(map[string]any)
			current[segment] = next
			current = next

			continue
		}

		next, isObject := existing.(map[string]any)
		if !isObject {
			return errConflictingFields
		}

		current = next
	}

	leaf := path[len(path)-1]
	if _, exists := current[leaf]; exists {
		return errConflictingFields
	}

	current[leaf] = value

	return nil
}

// selectEventFields whitelists the requested fields of an event record, resolving JSONPath
// fields (e.g. "$['start']['dateTime']") into nested maps the same way Gmail and Stripe do.
// Keys in the output are lowercased by the shared helper; "id" is always included.
func selectEventFields(record map[string]any, fields []string) map[string]any {
	return readhelper.SelectFields(record, datautils.NewSetFromList(append(fields, "id")))
}
