package common

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
var testJSONData = `{
		"count":38, "text":"Hello World", "pi":3.14, "metadata":null, "list":[1,2,3], "arr":[],
		"payload": {
			"notes": {
				"links": null,
				"body": {
					"text": "Some notes"
				}
			},
			"display_time": "yesterday"
		}}`

func TestJsonManager_GetInteger(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
	}

	tests := []struct {
		name        string
		input       inType
		expected    *int64
		expectedErr error
		withErr     bool
	}{
		{
			name: "Missing integer",
			input: inType{
				key: "age",
			},
			withErr: true,
		},
		{
			name: "Missing optional integer",
			input: inType{
				key:      "age",
				optional: true,
			},
			expected: nil,
		},
		{
			name: "Incorrect data type",
			input: inType{
				key: "text",
			},
			expectedErr: ErrNotNumeric,
		},
		{
			name: "Float is not integer",
			input: inType{
				key: "pi",
			},
			expectedErr: ErrNotInteger,
		},
		{
			name: "Key is found but Null",
			input: inType{
				key: "metadata",
			},
			expectedErr: ErrNullJSON,
		},
		{
			name: "Valid number",
			input: inType{
				key: "count",
			},
			expected: ptrInt64(38),
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := JSONManager.GetInteger(j, tt.input.key, tt.input.optional)
			assertJSONManagerOutput(t, tt.name, tt.expected, tt.expectedErr, tt.withErr, output, err)
		})
	}
}

func TestJsonManager_GetString(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
	}

	tests := []struct {
		name        string
		input       inType
		expected    *string
		expectedErr error
		withErr     bool
	}{
		{
			name: "Missing string",
			input: inType{
				key: "surname",
			},
			withErr: true,
		},
		{
			name: "Missing optional string",
			input: inType{
				key:      "surname",
				optional: true,
			},
			expected: nil,
		},
		{
			name: "Incorrect data type",
			input: inType{
				key: "count",
			},
			expectedErr: ErrNotString,
		},
		{
			name: "Key is found but Null",
			input: inType{
				key: "metadata",
			},
			expectedErr: ErrNullJSON,
		},
		{
			name: "Valid string",
			input: inType{
				key: "text",
			},
			expected: ptrString("Hello World"),
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := JSONManager.GetString(j, tt.input.key, tt.input.optional)
			assertJSONManagerOutput(t, tt.name, tt.expected, tt.expectedErr, tt.withErr, output, err)
		})
	}
}

func TestJsonManager_GetArr(t *testing.T) {
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	tests := []struct {
		name         string
		input        string
		expectedSize int
		expectedErr  error
		withErr      bool
	}{
		{
			name:         "Missing array",
			input:        "queue",
			expectedSize: 0,
			withErr:      true,
		},
		{
			name:        "Key is found but Null",
			input:       "metadata",
			expectedErr: ErrNotArray,
		},
		{
			name:         "Valid array",
			input:        "list",
			expectedSize: 3,
		},
		{
			name:         "Empty array",
			input:        "arr",
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := JSONManager.GetArr(j, tt.input)
			if tt.withErr {
				if err == nil {
					t.Fatalf("%s: expected error while none received", tt.name)
				}
			} else {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
				}

				outputSize := len(output)
				if !reflect.DeepEqual(outputSize, tt.expectedSize) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedSize, outputSize)
				}
			}
		})
	}
}

func TestJsonManager_GetNestedNode(t *testing.T) {
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	tests := []struct {
		name        string
		input       string
		expectedErr error
		withErr     bool
	}{
		{
			name:    "Missing nested object",
			input:   "payload.random",
			withErr: true,
		},
		{
			name:        "Nested key is found but Null",
			input:       "payload.notes.links",
			expectedErr: ErrNotObject,
		},
		{
			name:        "Invalid data type of existing key",
			input:       "payload.notes.body.text",
			expectedErr: ErrNotObject,
		},
		{
			name:  "Valid nested object",
			input: "payload.notes.body",
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := JSONManager.GetNestedNode(j, DotZoom(tt.input))
			if tt.withErr {
				if err == nil {
					t.Fatalf("%s: expected error while none received", tt.name)
				}
			} else {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
				}
			}
		})
	}
}

func helperCreateJSON(t *testing.T, text string) *ajson.Node {
	t.Helper()

	jsonBody, err := ajson.Unmarshal([]byte(text))
	if err != nil {
		t.Fatalf("bad test, JSON object is invalid, cannot proceed with the test")
	}

	return jsonBody
}

func assertJSONManagerOutput(t *testing.T, name string, expected any, expectedErr error,
	withErr bool, output any, err error,
) {
	t.Helper()

	if withErr {
		if err == nil {
			t.Fatalf("%s: expected error while none received", name)
		}
	} else {
		// check for actual error value
		if !errors.Is(err, expectedErr) {
			t.Fatalf("%s: expected: (%v), got: (%v)", name, expectedErr, err)
		}

		if !reflect.DeepEqual(output, expected) {
			t.Fatalf("%s: expected: (%v), got: (%v)", name, expected, output)
		}
	}
}

func ptrInt64(num int64) *int64 {
	return &num
}

func ptrString(text string) *string {
	return &text
}
