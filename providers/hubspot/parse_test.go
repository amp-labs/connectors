package hubspot

import (
	"context"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestGetDataMarshaller(t *testing.T) { //nolint:funlen
	t.Parallel()

	connector := &Connector{}
	marshaller := connector.getDataMarshaller(context.Background(), "contacts", nil)

	tests := []struct {
		name        string
		records     []map[string]any
		fields      []string
		expected    []common.ReadResultRow
		expectedErr error
	}{
		{
			name: "Fields from properties only",
			records: []map[string]any{
				{
					"id": "7462",
					"properties": map[string]any{
						"email":      "bruce@yahoo.com",
						"createdate": "2023-12-14T23:32:45.536Z",
					},
					"archived": false,
				},
			},
			fields: []string{"email"},
			expected: []common.ReadResultRow{
				{
					Id: "7462",
					Fields: map[string]any{
						"email": "bruce@yahoo.com",
					},
					Raw: map[string]any{
						"id": "7462",
						"properties": map[string]any{
							"email":      "bruce@yahoo.com",
							"createdate": "2023-12-14T23:32:45.536Z",
						},
						"archived": false,
					},
				},
			},
		},
		{
			name: "Top-level id field is included in result Fields",
			records: []map[string]any{
				{
					"id": "7462",
					"properties": map[string]any{
						"email":      "bruce@yahoo.com",
						"createdate": "2023-12-14T23:32:45.536Z",
					},
					"archived": false,
				},
			},
			fields: []string{"id", "email"},
			expected: []common.ReadResultRow{
				{
					Id: "7462",
					Fields: map[string]any{
						"id":    "7462",
						"email": "bruce@yahoo.com",
					},
					Raw: map[string]any{
						"id": "7462",
						"properties": map[string]any{
							"email":      "bruce@yahoo.com",
							"createdate": "2023-12-14T23:32:45.536Z",
						},
						"archived": false,
					},
				},
			},
		},
		{
			name: "Only id field requested",
			records: []map[string]any{
				{
					"id": "7462",
					"properties": map[string]any{
						"email": "bruce@yahoo.com",
					},
					"archived": false,
				},
			},
			fields: []string{"id"},
			expected: []common.ReadResultRow{
				{
					Id: "7462",
					Fields: map[string]any{
						"id": "7462",
					},
					Raw: map[string]any{
						"id": "7462",
						"properties": map[string]any{
							"email": "bruce@yahoo.com",
						},
						"archived": false,
					},
				},
			},
		},
		{
			name: "Properties field takes precedence over top-level field",
			records: []map[string]any{
				{
					"id": "7462",
					"properties": map[string]any{
						"id":    "properties-id",
						"email": "bruce@yahoo.com",
					},
				},
			},
			fields: []string{"id", "email"},
			expected: []common.ReadResultRow{
				{
					Id: "7462",
					Fields: map[string]any{
						"id":    "properties-id",
						"email": "bruce@yahoo.com",
					},
					Raw: map[string]any{
						"id": "7462",
						"properties": map[string]any{
							"id":    "properties-id",
							"email": "bruce@yahoo.com",
						},
					},
				},
			},
		},
		{
			name: "No fields requested leaves Fields nil",
			records: []map[string]any{
				{
					"id": "7462",
					"properties": map[string]any{
						"email": "bruce@yahoo.com",
					},
				},
			},
			fields: []string{},
			expected: []common.ReadResultRow{
				{
					Id:     "7462",
					Fields: nil,
					Raw: map[string]any{
						"id": "7462",
						"properties": map[string]any{
							"email": "bruce@yahoo.com",
						},
					},
				},
			},
		},
		{
			name: "Multiple records with id in fields",
			records: []map[string]any{
				{
					"id": "100",
					"properties": map[string]any{
						"email": "a@example.com",
					},
				},
				{
					"id": "200",
					"properties": map[string]any{
						"email": "b@example.com",
					},
				},
			},
			fields: []string{"id", "email"},
			expected: []common.ReadResultRow{
				{
					Id: "100",
					Fields: map[string]any{
						"id":    "100",
						"email": "a@example.com",
					},
					Raw: map[string]any{
						"id": "100",
						"properties": map[string]any{
							"email": "a@example.com",
						},
					},
				},
				{
					Id: "200",
					Fields: map[string]any{
						"id":    "200",
						"email": "b@example.com",
					},
					Raw: map[string]any{
						"id": "200",
						"properties": map[string]any{
							"email": "b@example.com",
						},
					},
				},
			},
		},
		{
			name:        "Missing id returns error",
			records:     []map[string]any{{"properties": map[string]any{}}},
			fields:      []string{"email"},
			expectedErr: errMissingId,
		},
		{
			name:        "Missing properties with fields requested returns error",
			records:     []map[string]any{{"id": "123"}},
			fields:      []string{"email"},
			expectedErr: ErrNotObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := marshaller(tt.records, tt.fields)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}

				if err != tt.expectedErr {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("result mismatch\ngot:  %+v\nwant: %+v", result, tt.expected)
			}
		})
	}
}

// nolint:funlen
func TestGetLastResultId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *common.ReadResultRow
		expected string
	}{
		{
			name:     "Nil Input",
			input:    nil,
			expected: "",
		},
		{
			name:     "Empty Data",
			input:    &common.ReadResultRow{},
			expected: "",
		},
		{
			name: "Valid Fields[id]",
			input: &common.ReadResultRow{
				Fields: map[string]any{string(ObjectFieldId): "12345"},
			},
			expected: "12345",
		},
		{
			name: "Valid Fields[hs_object_id]",
			input: &common.ReadResultRow{
				Fields: map[string]any{string(ObjectFieldHsObjectId): "67890"},
			},
			expected: "67890",
		},
		{
			name: "Valid Raw[id]",
			input: &common.ReadResultRow{
				Raw: map[string]any{string(ObjectFieldId): "abcdef"},
			},
			expected: "abcdef",
		},
		{
			name: "Valid Raw[properties][hs_object_id]",
			input: &common.ReadResultRow{
				Raw: map[string]any{
					string(ObjectFieldProperties): map[string]any{
						string(ObjectFieldHsObjectId): "ghijkl",
					},
				},
			},
			expected: "ghijkl",
		},
		{
			name: "Dummy hubspot test",
			input: &common.ReadResultRow{
				Fields: map[string]any{
					"lifecyclestage": "lead",
				},
				Raw: map[string]any{
					"archived":  false,
					"createdAt": "2010-12-08T06:13:17.698Z",
					"id":        "15237",
					"properties": map[string]any{
						"createdate":       "2010-12-08T06:13:17.698Z",
						"hs_object_id":     "15237",
						"lastmodifieddate": "2010-12-04T11:18:28.697Z",
						"lifecyclestage":   "lead",
					},
					"updatedAt": "2010-12-04T11:18:28.697Z",
				},
			},
			expected: "15237",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := GetResultId(test.input)
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}
