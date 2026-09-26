package objectcatalog

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestMetadataImmutability(t *testing.T) {
	// 1. MustLoadMetadata from JSON string
	// Static data will be shared across the test suite.
	// If implementation is correct it will be not modified.
	catalog := MustLoad([]byte((`
	{
		"users": {
			"displayName": "Users",
			"fields": {
				"id": {
					"displayName": "ID",
					"valueType": "string"
				}
			}
		}
	}`)))

	tests := []struct {
		name       string
		addFields  bool
		fieldToAdd string
	}{
		{
			name:      "No fields added",
			addFields: false,
		},
		{
			name:       "Add email field",
			addFields:  true,
			fieldToAdd: "email",
		},
		{
			name:       "Add occupation field",
			addFields:  true,
			fieldToAdd: "occupation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 2. SelectObjects
			selected, _ := catalog.Select([]string{"users"})

			// 3. AddField if required
			if tt.addFields {
				isCustom := true
				newField := common.FieldMetadata{
					DisplayName: "Email",
					ValueType:   "string",
					IsCustom:    &isCustom,
				}
				selected.AddField("users", tt.fieldToAdd, newField)
			}

			// 4. Verify results
			result := selected.Result()
			userMeta := result.GetObjectMetadata("users")

			comparison := testutils.NewCompareResult()

			// Check DisplayName
			comparison.Assert("DisplayName must be correct", "Users", userMeta.DisplayName)

			if tt.addFields {
				comparison.Assert("Selected metadata must have enhanced fields", 2, len(userMeta.Fields))
				if _, ok := userMeta.Fields[tt.fieldToAdd]; !ok {
					t.Errorf("expected %s field to be added to selected metadata", tt.fieldToAdd)
				}
				comparison.Assert("Selected metadata FieldsMap must have enhanced fields", 2, len(userMeta.FieldsMap))
				if displayName, ok := userMeta.FieldsMap[tt.fieldToAdd]; !ok || displayName != "Email" {
					t.Errorf("expected %s field to be in FieldsMap with correct DisplayName", tt.fieldToAdd)
				}
			} else {
				comparison.Assert("Selected metadata must have original fields", 1, len(userMeta.Fields))
				comparison.Assert("Selected metadata FieldsMap must have original fields", 1, len(userMeta.FieldsMap))
			}

			// 5. Verify original StaticMetadata is UNCHANGED
			originalUserMeta := catalog.Objects["users"]
			comparison.Assert("Original StaticMetadata must remain unchanged", 1, len(originalUserMeta.Fields))
			if _, ok := originalUserMeta.Fields["email"]; ok {
				t.Errorf("StaticMetadata was modified! AddField should not affect it.")
			}

			comparison.Validate(t, tt.name)
		})
	}
}
