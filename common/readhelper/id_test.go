// nolint
package readhelper

import (
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdField(t *testing.T) {
	t.Parallel()

	query := NewIdField("id")

	assert.Equal(t, "id", query.Field)
	assert.Nil(t, query.Zoom)
}

func TestNewNestedIdField(t *testing.T) {
	t.Parallel()

	query := NewNestedIdField([]string{"meta", "info"}, "uid")

	assert.Equal(t, "uid", query.Field)
	assert.Equal(t, []string{"meta", "info"}, query.Zoom)
}

func TestExtractIdFromRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		record   map[string]any
		query    IdFieldQuery
		expected string
	}{
		{
			name:     "flat string id",
			record:   map[string]any{"id": "123", "name": "test"},
			query:    NewIdField("id"),
			expected: "123",
		},
		{
			name:     "flat numeric id (float64)",
			record:   map[string]any{"id": float64(456), "name": "test"},
			query:    NewIdField("id"),
			expected: "456",
		},
		{
			name:     "nested id one level deep",
			record:   map[string]any{"id": map[string]any{"record_id": "nested-123"}},
			query:    NewNestedIdField([]string{"id"}, "record_id"),
			expected: "nested-123",
		},
		{
			name: "deeply nested id",
			record: map[string]any{
				"meta": map[string]any{
					"info": map[string]any{
						"uid": "deep-456",
					},
				},
			},
			query:    NewNestedIdField([]string{"meta", "info"}, "uid"),
			expected: "deep-456",
		},
		{
			name:     "missing id field returns empty string",
			record:   map[string]any{"name": "test"},
			query:    NewIdField("id"),
			expected: "",
		},
		{
			name:     "missing nested path returns empty string",
			record:   map[string]any{"id": "123"},
			query:    NewNestedIdField([]string{"nonexistent"}, "record_id"),
			expected: "",
		},
		{
			name:     "nested path is not an object returns empty string",
			record:   map[string]any{"id": "not-an-object"},
			query:    NewNestedIdField([]string{"id"}, "record_id"),
			expected: "",
		},
		{
			name:     "unsupported id type returns empty string",
			record:   map[string]any{"id": []string{"array-value"}},
			query:    NewIdField("id"),
			expected: "",
		},
		{
			name:     "custom field name",
			record:   map[string]any{"userId": "user-789", "name": "test"},
			query:    NewIdField("userId"),
			expected: "user-789",
		},
		{
			name:     "numeric id with decimals truncated",
			record:   map[string]any{"id": float64(123.999)},
			query:    NewIdField("id"),
			expected: "124",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractIdFromRecord(tt.record, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMakeGetMarshaledDataWithId(t *testing.T) {
	t.Parallel()

	// Create a default map where most objects use "id" field
	idMapping := datautils.NewDefaultMap(
		datautils.Map[string, IdFieldQuery]{
			"specialObject": NewNestedIdField([]string{"id"}, "record_id"),
		},
		func(_ string) IdFieldQuery {
			return NewIdField("id")
		},
	)

	tests := []struct {
		name           string
		objectName     string
		records        []map[string]any
		fields         []string
		expectedIds    []string
		expectedFields []string
	}{
		{
			name:       "standard flat id extraction",
			objectName: "contacts",
			records: []map[string]any{
				{"id": "1", "name": "Alice", "email": "alice@example.com"},
				{"id": "2", "name": "Bob", "email": "bob@example.com"},
			},
			fields:         []string{"name", "email"},
			expectedIds:    []string{"1", "2"},
			expectedFields: []string{"name", "email"},
		},
		{
			name:       "special object with nested id",
			objectName: "specialObject",
			records: []map[string]any{
				{"id": map[string]any{"record_id": "nested-1"}, "name": "Item1"},
				{"id": map[string]any{"record_id": "nested-2"}, "name": "Item2"},
			},
			fields:         []string{"name"},
			expectedIds:    []string{"nested-1", "nested-2"},
			expectedFields: []string{"name"},
		},
		{
			name:       "empty records returns empty slice",
			objectName: "contacts",
			records:    []map[string]any{},
			fields:     []string{"name"},
			expectedIds: []string{},
		},
		{
			name:       "missing id gracefully returns empty string",
			objectName: "contacts",
			records: []map[string]any{
				{"name": "No ID Record"},
			},
			fields:         []string{"name"},
			expectedIds:    []string{""},
			expectedFields: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			marshalFunc := MakeGetMarshaledDataWithId(tt.objectName, idMapping)
			result, err := marshalFunc(tt.records, tt.fields)

			require.NoError(t, err)
			assert.Len(t, result, len(tt.records))

			for i, row := range result {
				if i < len(tt.expectedIds) {
					assert.Equal(t, tt.expectedIds[i], row.Id)
				}
			}
		})
	}
}

func TestMakeMarshaledDataFuncWithId(t *testing.T) {
	t.Parallel()

	// Create a default map where most objects use "id" field
	idMapping := datautils.NewDefaultMap(
		datautils.Map[string, IdFieldQuery]{},
		func(_ string) IdFieldQuery {
			return NewIdField("id")
		},
	)

	createTestNode := func(jsonStr string) *ajson.Node {
		node, err := ajson.Unmarshal([]byte(jsonStr))
		require.NoError(t, err)

		return node
	}

	tests := []struct {
		name        string
		objectName  string
		records     []*ajson.Node
		fields      []string
		expectedIds []string
	}{
		{
			name:       "extracts id from ajson nodes",
			objectName: "contacts",
			records: []*ajson.Node{
				createTestNode(`{"id": "node-1", "name": "Alice"}`),
				createTestNode(`{"id": "node-2", "name": "Bob"}`),
			},
			fields:      []string{"name"},
			expectedIds: []string{"node-1", "node-2"},
		},
		{
			name:        "empty records returns empty slice",
			objectName:  "contacts",
			records:     []*ajson.Node{},
			fields:      []string{"name"},
			expectedIds: []string{},
		},
		{
			name:       "numeric id converted to string",
			objectName: "contacts",
			records: []*ajson.Node{
				createTestNode(`{"id": 12345, "name": "Test"}`),
			},
			fields:      []string{"name"},
			expectedIds: []string{"12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			marshalFunc := MakeMarshaledDataFuncWithId(nil, tt.objectName, idMapping)
			result, err := marshalFunc(tt.records, tt.fields)

			require.NoError(t, err)
			assert.Len(t, result, len(tt.records))

			for i, row := range result {
				if i < len(tt.expectedIds) {
					assert.Equal(t, tt.expectedIds[i], row.Id)
				}
			}
		})
	}
}

func TestMakeMarshaledDataFuncWithId_WithTransformer(t *testing.T) {
	t.Parallel()

	idMapping := datautils.NewDefaultMap(
		datautils.Map[string, IdFieldQuery]{},
		func(_ string) IdFieldQuery {
			return NewIdField("id")
		},
	)

	// Custom transformer that flattens "attributes" into root
	transformer := func(node *ajson.Node) (map[string]any, error) {
		result := make(map[string]any)

		if idNode := node.MustKey("id"); idNode != nil {
			result["id"], _ = idNode.GetString()
		}

		if attrs := node.MustKey("attributes"); attrs != nil {
			if nameNode := attrs.MustKey("name"); nameNode != nil {
				result["name"], _ = nameNode.GetString()
			}
		}

		return result, nil
	}

	jsonStr := `{"id": "transformed-1", "attributes": {"name": "TransformedName"}}`
	node, err := ajson.Unmarshal([]byte(jsonStr))
	require.NoError(t, err)

	marshalFunc := MakeMarshaledDataFuncWithId(transformer, "contacts", idMapping)
	result, err := marshalFunc([]*ajson.Node{node}, []string{"name"})

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "transformed-1", result[0].Id)
	assert.Equal(t, "TransformedName", result[0].Fields["name"])
}
