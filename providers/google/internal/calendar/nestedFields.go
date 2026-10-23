package calendar

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Google Calendar events keep their timing, origin and tracing data inside nested objects.
// The metadata schema exposes those leaves in JSONPath notation, e.g. "$['start']['dateTime']",
// following the convention used by Gmail. Bracket notation keeps keys that contain dots
// (allowed in extendedProperties, e.g. "com.acme.dealId") as a single segment.
// https://developers.google.com/workspace/calendar/api/v3/reference/events#resource

var (
	errConflictingFields         = errors.New("conflicting fields in record data")
	errEmptyFieldPath            = errors.New("empty field path in record data")
	errArrayWildcardNotSupported = errors.New("array wildcards are not supported when writing")
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

// recordKey is a record data key with its parsed JSONPath. A plain key, or a single-segment
// JSONPath such as "$['summary']", has a one-element path holding the provider field name.
type recordKey struct {
	key  string
	path []string
}

// nestRecordData converts JSONPath keys in record data into nested objects, so that
// {"$['start']['dateTime']": "..."} becomes {"start": {"dateTime": "..."}}.
//
// Plain keys are applied before JSONPath keys, so a nested object the caller supplied is in
// place for JSONPath keys to merge into; within each group keys are sorted, so the result is
// deterministic. The caller's maps are never modified. It is an error when two keys overlap,
// e.g. "start" holding a string and "$['start']['dateTime']" both present, or when the same
// leaf is set twice.
func nestRecordData(data map[string]any) (map[string]any, error) {
	keys := make([]recordKey, 0, len(data))
	for key := range data {
		keys = append(keys, recordKey{key: key, path: readhelper.ParseJSONPath(key)})
	}

	slices.SortFunc(keys, func(a, b recordKey) int {
		if plainA, plainB := len(a.path) == 1, len(b.path) == 1; plainA != plainB {
			if plainA {
				return -1
			}

			return 1
		}

		return strings.Compare(a.key, b.key)
	})

	result := make(map[string]any, len(data))

	for _, item := range keys {
		if err := setNestedValue(result, item.path, data[item.key]); err != nil {
			return nil, fmt.Errorf("%w: %v", err, item.key)
		}
	}

	return result, nil
}

func setNestedValue(target map[string]any, path []string, value any) error {
	if slices.Contains(path, readhelper.ArrayWildIndex) {
		return errArrayWildcardNotSupported
	}

	if len(path) == 0 || path[0] == "" {
		return errEmptyFieldPath
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

		// Copy on write: nested objects copied from the caller's record data must not be
		// mutated, otherwise a retry of the same write would see its own leaf and conflict.
		next = maps.Clone(next)
		current[segment] = next
		current = next
	}

	leaf := path[len(path)-1]
	if _, exists := current[leaf]; exists {
		return errConflictingFields
	}

	current[leaf] = value

	return nil
}

// eventFieldSelection is a requested field list split once into plain fields and
// multi-segment JSONPath fields, so the split is not repeated for every record on a page.
type eventFieldSelection struct {
	plain  []string
	nested []string
}

// newEventFieldSelection classifies requested fields. "id" is always included. A
// single-segment JSONPath such as "$['summary']" counts as the plain field "summary".
func newEventFieldSelection(fields []string) eventFieldSelection {
	selection := eventFieldSelection{plain: []string{"id"}}

	for _, field := range fields {
		if path := readhelper.ParseJSONPath(field); len(path) == 1 {
			selection.plain = append(selection.plain, path[0])
		} else {
			selection.nested = append(selection.nested, field)
		}
	}

	return selection
}

// apply whitelists the requested fields of an event record. Output keys are lowercased, as
// for every other object.
//
// Plain fields are matched case-insensitively, preserving the behaviour events had before
// nested fields existed. JSONPath fields are resolved into nested maps by the shared
// selector, the way Gmail and Stripe do. When a parent and one of its children are both
// requested the parent wins, so the result does not depend on iteration order.
//
// The shared selector may write into the map it selects from in that parent-and-child case,
// so callers must not reuse the record as Raw.
func (s eventFieldSelection) apply(record map[string]any) map[string]any {
	result := common.ExtractLowercaseFieldsFromRaw(s.plain, record)

	if len(s.nested) == 0 {
		return result
	}

	for key, value := range readhelper.SelectFields(record, datautils.NewSetFromList(s.nested)) {
		if _, exists := result[key]; !exists {
			result[key] = value
		}
	}

	return result
}
