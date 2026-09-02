package openapi

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSoleNonNullMember(t *testing.T) {
	t.Parallel()

	stringSchema := map[string]any{"type": "string"}
	intSchema := map[string]any{"type": "integer"}
	nullSchema := map[string]any{"type": "null"}
	refSchema := map[string]any{"$ref": "#/components/schemas/ScoringInfo"}

	tests := []struct {
		name      string
		input     any
		want      map[string]any
		wantFound bool
	}{{
		name:      "nullable string collapses to the string",
		input:     []any{stringSchema, nullSchema},
		want:      stringSchema,
		wantFound: true,
	}, {
		name:      "order does not matter",
		input:     []any{nullSchema, stringSchema},
		want:      stringSchema,
		wantFound: true,
	}, {
		name:      "a $ref is a valid survivor",
		input:     []any{refSchema, nullSchema},
		want:      refSchema,
		wantFound: true,
	}, {
		name:      "several nulls still collapse",
		input:     []any{nullSchema, stringSchema, nullSchema},
		want:      stringSchema,
		wantFound: true,
	}, {
		// The whole point of the sawNull condition: a genuine union carries
		// information, so "other" is the honest answer and it must be left alone.
		name:      "union of two real types is left alone",
		input:     []any{stringSchema, intSchema},
		wantFound: false,
	}, {
		name:      "union of two real types plus a null is still ambiguous",
		input:     []any{stringSchema, intSchema, nullSchema},
		wantFound: false,
	}, {
		// Without a null there was no optionality to strip, so rewriting would
		// be a change in meaning rather than a simplification.
		name:      "lone non-null member without a null is left alone",
		input:     []any{stringSchema},
		wantFound: false,
	}, {
		name:      "all-null list has no survivor",
		input:     []any{nullSchema},
		wantFound: false,
	}, {
		name:      "empty list",
		input:     []any{},
		wantFound: false,
	}, {
		name:      "non-object member disqualifies the whole list",
		input:     []any{stringSchema, nullSchema, "garbage"},
		wantFound: false,
	}, {
		name:      "non-slice input",
		input:     map[string]any{"type": "string"},
		wantFound: false,
	}, {
		name:      "nil input",
		input:     nil,
		wantFound: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, found := soleNonNullMember(tt.input)

			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollapseNullableAnyOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{{
		// Monaco marks nearly every response field this way. Without the
		// rewrite the extractor sees no `type` and reports valueType "other".
		name:  "nullable field gains the inner type",
		input: `{"anyOf":[{"type":"string"},{"type":"null"}],"title":"Email"}`,
		want:  `{"title":"Email","type":"string"}`,
	}, {
		name:  "parent keys win over the survivor's",
		input: `{"anyOf":[{"type":"string","title":"Inner"},{"type":"null"}],"title":"Outer"}`,
		want:  `{"title":"Outer","type":"string"}`,
	}, {
		name:  "genuine union is untouched",
		input: `{"anyOf":[{"type":"string"},{"type":"integer"}],"title":"Mixed"}`,
		want:  `{"anyOf":[{"type":"string"},{"type":"integer"}],"title":"Mixed"}`,
	}, {
		name: "nested properties are rewritten too",
		input: `{"components":{"schemas":{"Contact":{"properties":{` +
			`"email":{"anyOf":[{"type":"string"},{"type":"null"}]}}}}}}`,
		want: `{"components":{"schemas":{"Contact":{"properties":{` +
			`"email":{"type":"string"}}}}}}`,
	}, {
		name:  "arrays of schemas are walked",
		input: `[{"anyOf":[{"type":"string"},{"type":"null"}]}]`,
		want:  `[{"type":"string"}]`,
	}, {
		name:  "schema without anyOf is unchanged",
		input: `{"type":"boolean","default":false}`,
		want:  `{"default":false,"type":"boolean"}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := collapseNullableAnyOf([]byte(tt.input))

			if !sameJSON(t, got, []byte(tt.want)) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestCollapseNullableAnyOfInvalidInput asserts that unparseable input is
// handed back untouched, so a malformed spec surfaces as the explorer's parse
// error rather than being silently replaced with something else.
func TestCollapseNullableAnyOfInvalidInput(t *testing.T) {
	t.Parallel()

	input := []byte(`{"not":"valid`)

	if got := collapseNullableAnyOf(input); !reflect.DeepEqual(got, input) {
		t.Errorf("got %s, want the input returned verbatim", got)
	}
}

func sameJSON(t *testing.T, left, right []byte) bool {
	t.Helper()

	var leftValue, rightValue any

	if err := json.Unmarshal(left, &leftValue); err != nil {
		t.Fatalf("left is not JSON: %v", err)
	}

	if err := json.Unmarshal(right, &rightValue); err != nil {
		t.Fatalf("right is not JSON: %v", err)
	}

	return reflect.DeepEqual(leftValue, rightValue)
}
