package associations

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func TestGetDataMarshaller(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name              string
		records           []map[string]any
		fields            []string
		associatedObjects []string
		expected          []common.ReadResultRow
		expectedErr       error
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
			expectedErr: jsonquery.ErrKeyNotFound,
		},
		{
			name:        "Missing properties with fields requested returns error",
			records:     []map[string]any{{"id": "123"}},
			fields:      []string{"email"},
			expectedErr: jsonquery.ErrKeyNotFound,
		},
		{
			name: "Several records with associations",
			records: []map[string]any{{
				"id": "356",
				"properties": map[string]any{
					"email": "one@example.com",
				},
			}, {
				"id": "772",
				"properties": map[string]any{
					"email": "two@example.com",
				},
			}},
			fields:            []string{"email"},
			associatedObjects: []string{"deals"},
			expected: []common.ReadResultRow{{
				Id: "356",
				Fields: map[string]any{
					"email": "one@example.com",
				},
				Raw: map[string]any{
					"id": "356",
					"properties": map[string]any{
						"email": "one@example.com",
					},
				},
				Associations: map[string][]common.Association{
					"deals": {{
						ObjectId:        "assoc-356",
						AssociationType: "test-type",
					}},
				},
			}, {
				Id: "772",
				Fields: map[string]any{
					"email": "two@example.com",
				},
				Raw: map[string]any{
					"id": "772",
					"properties": map[string]any{
						"email": "two@example.com",
					},
				},
				Associations: map[string][]common.Association{
					"deals": {{
						ObjectId:        "assoc-772",
						AssociationType: "test-type",
					}},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			marshaller := CreateDataMarshallerWithAssociations(
				context.Background(),
				&testFiller{},
				"contacts",
				tt.associatedObjects,
			)

			records, err := datautils.ForEachWithErr(tt.records, func(object map[string]any) (*ajson.Node, error) {
				return jsonquery.Convertor.NodeFromMap(object)
			})

			result, err := marshaller(records, tt.fields)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}

				if !errors.Is(err, tt.expectedErr) {
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

// associationsFiller is a test stub that overrides the default
// association-filling behavior of the AssociationsFiller.
type testFiller struct{}

func (testFiller) FillAssociations(
	ctx context.Context, fromObjName string, toAssociatedObjects []string,
	data []common.ReadResultRow,
) error {
	if len(toAssociatedObjects) == 0 {
		return nil // nothing to do
	}

	for index, row := range data {
		if data[index].Associations == nil {
			data[index].Associations = make(map[string][]common.Association)
		}

		for _, toObj := range toAssociatedObjects {
			data[index].Associations[toObj] = []common.Association{{
				ObjectId:        "assoc-" + row.Id,
				AssociationType: "test-type",
			}}
		}
	}

	return nil
}
