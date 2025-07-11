package zohocrm

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestBuildNestedFieldGroups(t *testing.T) {
	t.Parallel()

	// Test data
	watchFieldsMetadata := map[string]map[string]any{
		"field1": {"id": "1", "api_name": "Field1"},
		"field2": {"id": "2", "api_name": "Field2"},
		"field3": {"id": "3", "api_name": "Field3"},
		"field4": {"id": "4", "api_name": "Field4"},
	}

	fieldNames := []string{"field1", "field2", "field3", "field4"}

	// Test the nested field groups
	result, err := recursiveFieldSelectionBuild(fieldNames, watchFieldsMetadata)
	if err != nil {
		t.Fatalf("buildFieldSelection failed: %v", err)
	}

	// Top-level should have 2 groups: field1 and a nested group
	assert.Assert(t, result.Group != nil)
	assert.Equal(t, string(result.GroupOperator), string(GroupOperatorOr))
	assert.Equal(t, len(result.Group), 2)

	// First group should be field1
	assert.Assert(t, result.Group[0].Field != nil)
	assert.Equal(t, result.Group[0].Field.APIName, "Field1")
	assert.Equal(t, result.Group[0].Field.ID, "1")

	// Second group should be a nested group for field2, field3, field4
	nested := result.Group[1]
	assert.Assert(t, nested.Group != nil)
	assert.Equal(t, len(nested.Group), 2)
	assert.Equal(t, nested.GroupOperator, string(GroupOperatorOr))

	// nested.Group[0] should be field2
	assert.Assert(t, nested.Group[0].Field != nil)
	assert.Equal(t, nested.Group[0].Field.APIName, "Field2")
	assert.Equal(t, nested.Group[0].Field.ID, "2")

	// nested.Group[1] should be a nested group for field3, field4
	nested2 := nested.Group[1]
	assert.Assert(t, nested2.Group != nil)
	assert.Equal(t, len(nested2.Group), 2)
	assert.Equal(t, nested2.GroupOperator, string(GroupOperatorOr))
	assert.Equal(t, nested2.Group[0].Field.APIName, "Field3")
	assert.Equal(t, nested2.Group[0].Field.ID, "3")
	assert.Equal(t, nested2.Group[1].Field.APIName, "Field4")
	assert.Equal(t, nested2.Group[1].Field.ID, "4")
}

func TestBuildNestedFieldGroupsEdgeCases(t *testing.T) {
	t.Parallel()

	// Test with 1 field
	watchFieldsMetadata1 := map[string]map[string]any{
		"field1": {"id": "1", "api_name": "Field1"},
	}
	fieldNames1 := []string{"field1"}

	result1, err := recursiveFieldSelectionBuild(fieldNames1, watchFieldsMetadata1)
	if err != nil {
		t.Fatalf("buildFieldSelection failed: %v", err)
	}

	assert.Assert(t, result1.Field != nil)
	assert.Equal(t, result1.Field.APIName, "Field1")
	assert.Equal(t, result1.Field.ID, "1")

	// Test with 2 fields
	watchFieldsMetadata2 := map[string]map[string]any{
		"field1": {"id": "1", "api_name": "Field1"},
		"field2": {"id": "2", "api_name": "Field2"},
	}
	fieldNames2 := []string{"field1", "field2"}

	result2, err := recursiveFieldSelectionBuild(fieldNames2, watchFieldsMetadata2)
	if err != nil {
		t.Fatalf("buildFieldSelection failed: %v", err)
	}

	assert.Assert(t, result2.Group != nil)
	assert.Equal(t, string(result2.GroupOperator), string(GroupOperatorOr))
	assert.Equal(t, len(result2.Group), 2)
	assert.Equal(t, result2.Group[0].Field.APIName, "Field1")
	assert.Equal(t, result2.Group[1].Field.APIName, "Field2")

	// Test with empty fields
	result3, err := recursiveFieldSelectionBuild([]string{}, map[string]map[string]any{})
	if err != nil {
		t.Fatalf("buildFieldSelection failed: %v", err)
	}

	assert.Assert(t, result3.Field == nil)
	assert.Assert(t, result3.Group == nil)
}

func TestFormatAPIName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected string
	}{
		{"first_name", "First_Name"},
		{"last_name", "Last_Name"},
		{"phone", "Phone"},
		{"company", "Company"},
		{"industry", "Industry"},
	}

	watchFieldsMetadata := map[string]map[string]any{
		"first_name": {"id": "1", "api_name": "First_Name"},
		"last_name":  {"id": "2", "api_name": "Last_Name"},
		"phone":      {"id": "3", "api_name": "Phone"},
		"company":    {"id": "4", "api_name": "Company"},
		"industry":   {"id": "5", "api_name": "Industry"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			result, err := formatAPIName(tc.input, watchFieldsMetadata)
			if err != nil {
				t.Fatalf("formatAPIName failed: %v", err)
			}

			assert.Equal(t, result, tc.expected)
		})
	}
}
