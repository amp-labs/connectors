package calendar

import (
	"strings"

	"github.com/amp-labs/connectors/common"
)

// nestedFieldSeparator joins the segments of a nested field path, e.g. "start.dateTime".
//
// Google Calendar events keep their timing, origin and tracing data inside nested objects
// (start, end, source, extendedProperties). The metadata schema exposes the useful leaves
// as dotted field names so that they can be mapped individually. Those dotted names are
// translated back into nested JSON on write and resolved from nested JSON on read.
//
// https://developers.google.com/workspace/calendar/api/v3/reference/events#resource
const nestedFieldSeparator = "."

// nestRecordData converts dotted keys in record data into nested objects, so that
// {"start.dateTime": "..."} becomes {"start": {"dateTime": "..."}}.
//
// Dotted keys are merged into an existing nested object when the caller also supplied one.
// A non-object value already stored at an intermediate path is replaced by an object, since
// the provider would reject it anyway. Keys without a separator are copied as-is.
func nestRecordData(data map[string]any) map[string]any {
	result := make(map[string]any, len(data))

	// Plain keys first so that dotted keys merge into any nested object the caller sent.
	for key, value := range data {
		if !strings.Contains(key, nestedFieldSeparator) {
			result[key] = value
		}
	}

	for key, value := range data {
		if !strings.Contains(key, nestedFieldSeparator) {
			continue
		}

		segments := strings.Split(key, nestedFieldSeparator)
		setNestedValue(result, segments, value)
	}

	return result
}

func setNestedValue(target map[string]any, segments []string, value any) {
	current := target

	for _, segment := range segments[:len(segments)-1] {
		next, ok := current[segment].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[segment] = next
		}

		current = next
	}

	current[segments[len(segments)-1]] = value
}

// marshalRecordsWithNestedFields behaves like common.GetMarshaledData, but additionally
// resolves requested dotted field names (e.g. "start.dateTime") against the nested raw
// record. The resolved value is placed into ReadResultRow.Fields under the lowercased dotted
// name, matching how top-level fields are keyed. Raw is left untouched.
func marshalRecordsWithNestedFields(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	rows, err := common.GetMarshaledData(records, fields)
	if err != nil {
		return nil, err
	}

	for index := range rows {
		for _, field := range fields {
			if !strings.Contains(field, nestedFieldSeparator) {
				continue
			}

			if value, ok := lookupNestedValue(rows[index].Raw, strings.Split(field, nestedFieldSeparator)); ok {
				rows[index].Fields[strings.ToLower(field)] = value
			}
		}
	}

	return rows, nil
}

// lookupNestedValue walks the record following path segments. Each segment is matched
// case-insensitively, consistent with how top-level field selection works.
func lookupNestedValue(record map[string]any, segments []string) (any, bool) {
	var current any = record

	for _, segment := range segments {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		current, ok = lookupKeyFold(object, segment)
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func lookupKeyFold(object map[string]any, key string) (any, bool) {
	if value, ok := object[key]; ok {
		return value, true
	}

	for candidate, value := range object {
		if strings.EqualFold(candidate, key) {
			return value, true
		}
	}

	return nil, false
}
