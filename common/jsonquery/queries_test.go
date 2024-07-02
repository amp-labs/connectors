package jsonquery

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
var testJSONData = `{
		"count":38, "text":"Hello World", "pi":3.14, "metadata":null, "list":[1,2,3], "arr":[],
		"inProgress": false,
		"payload": {
			"notes": {
				"links": null,
				"body": {
					"text": "Some notes",
					"amount": 359,
					"purchased": true
				},
				"nested_arr": [1,2,3,4,5]
			},
			"display_time": "yesterday"
		}}`

func TestQuery_Integer(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
		zoom     []string
	}

	tests := []struct {
		name        string
		input       inType
		expected    *int64
		expectedErr error
	}{
		{
			name: "Missing required integer",
			input: inType{
				key:      "age",
				optional: false,
			},
			expectedErr: ErrKeyNotFound,
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
		{
			name: "Reaching for nested integer",
			input: inType{
				key:      "amount",
				optional: false,
				zoom:     []string{"payload", "notes", "body"},
			},
			expected: ptrInt64(359),
		},
		{
			name: "Non existent zoom path is ok for optional integer",
			input: inType{
				key:      "street",
				optional: true,
				zoom:     []string{"payload", "location", "address"},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := New(j, tt.input.zoom...).Integer(tt.input.key, tt.input.optional)
			assertJSONManagerOutput(t, tt.name, tt.expected, tt.expectedErr, output, err)
		})
	}
}

func TestQueryString(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
		zoom     []string
	}

	tests := []struct {
		name        string
		input       inType
		expected    *string
		expectedErr error
	}{
		{
			name: "Missing required string",
			input: inType{
				key:      "surname",
				optional: false,
			},
			expectedErr: ErrKeyNotFound,
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
		{
			name: "Reaching for nested string",
			input: inType{
				key:      "text",
				optional: false,
				zoom:     []string{"payload", "notes", "body"},
			},
			expected: ptrString("Some notes"),
		},
		{
			name: "Non existent zoom path is ok for optional string",
			input: inType{
				key:      "street",
				optional: true,
				zoom:     []string{"payload", "location", "address"},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := New(j, tt.input.zoom...).Str(tt.input.key, tt.input.optional)
			assertJSONManagerOutput(t, tt.name, tt.expected, tt.expectedErr, output, err)
		})
	}
}

func TestQueryBool(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
		zoom     []string
	}

	tests := []struct {
		name        string
		input       inType
		expected    *bool
		expectedErr error
	}{
		{
			name: "Missing required bool",
			input: inType{
				key:      "completed",
				optional: false,
			},
			expectedErr: ErrKeyNotFound,
		},
		{
			name: "Missing optional bool",
			input: inType{
				key:      "completed",
				optional: true,
			},
			expected: nil,
		},
		{
			name: "Incorrect data type",
			input: inType{
				key: "count",
			},
			expectedErr: ErrNotBool,
		},
		{
			name: "Key is found but Null",
			input: inType{
				key: "metadata",
			},
			expectedErr: ErrNullJSON,
		},
		{
			name: "Valid Bool",
			input: inType{
				key: "inProgress",
			},
			expected: ptrBool(false),
		},
		{
			name: "Reaching for nested bool",
			input: inType{
				key:      "purchased",
				optional: false,
				zoom:     []string{"payload", "notes", "body"},
			},
			expected: ptrBool(true),
		},
		{
			name: "Non existent zoom path is ok for optional bool",
			input: inType{
				key:      "applicable",
				optional: true,
				zoom:     []string{"payload", "location", "address"},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := New(j, tt.input.zoom...).Bool(tt.input.key, tt.input.optional)
			assertJSONManagerOutput(t, tt.name, tt.expected, tt.expectedErr, output, err)
		})
	}
}

func TestQuery_Array(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
		zoom     []string
	}

	tests := []struct {
		name         string
		input        inType
		expectedSize int
		expectedErr  error
	}{
		{
			name: "Missing required array",
			input: inType{
				key:      "queue",
				optional: false,
			},
			expectedSize: 0,
			expectedErr:  ErrKeyNotFound,
		},
		{
			name: "Key is found but Null",
			input: inType{
				key:      "metadata",
				optional: false,
			},
			expectedErr: ErrNullJSON,
		},
		{
			name: "Value with wrong type",
			input: inType{
				key:      "text",
				optional: false,
			},
			expectedErr: ErrNotArray,
		},
		{
			name: "Valid array",
			input: inType{
				key:      "list",
				optional: false,
			},
			expectedSize: 3,
		},
		{
			name: "Empty array",
			input: inType{
				key:      "arr",
				optional: false,
			},
			expectedSize: 0,
			expectedErr:  nil,
		},
		{
			name: "Reaching for nested array",
			input: inType{
				key:      "nested_arr",
				optional: false,
				zoom:     []string{"payload", "notes"},
			},
			expectedSize: 5,
		},
		{
			name: "Non existent zoom path is ok for optional arr",
			input: inType{
				key:      "street",
				optional: true,
				zoom:     []string{"payload", "location", "address"},
			},
			expectedSize: 0,
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := New(j, tt.input.zoom...).Array(tt.input.key, tt.input.optional)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
			}

			outputSize := len(output)
			if !reflect.DeepEqual(outputSize, tt.expectedSize) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedSize, outputSize)
			}
		})
	}
}

func TestQuery_Object(t *testing.T) { // nolint:funlen
	t.Parallel()

	j := helperCreateJSON(t, testJSONData) // nolint:varnamelen

	type inType struct {
		key      string
		optional bool
		zoom     []string
	}

	tests := []struct {
		name        string
		input       inType
		expectedErr error
	}{
		{
			name: "Missing required nested object",
			input: inType{
				key:  "random",
				zoom: []string{"payload"},
			},
			expectedErr: ErrKeyNotFound,
		},
		{
			name: "Nested key is found but Null",
			input: inType{
				key:  "links",
				zoom: []string{"payload", "notes"},
			},
			expectedErr: ErrNullJSON,
		},
		{
			name: "Invalid data type of existing key",
			input: inType{
				key:  "text",
				zoom: []string{"payload", "notes", "body"},
			},
			expectedErr: ErrNotObject,
		},
		{
			name: "Valid nested object",
			input: inType{
				key:  "body",
				zoom: []string{"payload", "notes"},
			},
			expectedErr: nil, // success
		},
		{
			name: "Non existent zoom path is ok for optional object",
			input: inType{
				key:      "street",
				optional: true,
				zoom:     []string{"payload", "location", "address"},
			},
			expectedErr: nil, // success
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(j, tt.input.zoom...).Object(tt.input.key, tt.input.optional)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
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
	output any, err error,
) {
	t.Helper()

	// check for actual error value
	if !errors.Is(err, expectedErr) {
		t.Fatalf("%s: expected: (%v), got: (%v)", name, expectedErr, err)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("%s: expected: (%v), got: (%v)", name, expected, output)
	}
}

func ptrInt64(num int64) *int64 {
	return &num
}

func ptrString(text string) *string {
	return &text
}

func ptrBool(flag bool) *bool {
	return &flag
}
