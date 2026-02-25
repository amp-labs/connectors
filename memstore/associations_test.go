package memstore

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test schemas with different association types
const (
	accountSchemaWithRaw = `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"x-amp-id-field": true
			},
			"name": {
				"type": "string"
			},
			"industry": {
				"type": "string"
			}
		},
		"required": ["name"]
	}`

	contactSchemaWithForeignKey = `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"x-amp-id-field": true
			},
			"name": {
				"type": "string"
			},
			"email": {
				"type": "string"
			},
			"account_id": {
				"type": "string",
				"x-amp-association": {
					"associationType": "foreignKey",
					"targetObject": "account"
				}
			}
		},
		"required": ["name", "email"]
	}`

	opportunitySchemaWithReverseLookup = `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"x-amp-id-field": true
			},
			"name": {
				"type": "string"
			},
			"account_id": {
				"type": "string"
			},
			"contacts": {
				"type": "array",
				"x-amp-association": {
					"associationType": "reverseLookup",
					"targetObject": "opportunityContact",
					"foreignKeyField": "opportunity_id"
				}
			}
		},
		"required": ["name"]
	}`

	opportunityContactSchema = `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"x-amp-id-field": true
			},
			"opportunity_id": {
				"type": "string"
			},
			"contact_id": {
				"type": "string"
			}
		},
		"required": ["opportunity_id", "contact_id"]
	}`

	dealSchemaWithJunction = `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"x-amp-id-field": true
			},
			"name": {
				"type": "string"
			},
			"contacts": {
				"type": "array",
				"x-amp-association": {
					"associationType": "junction",
					"targetObject": "contact",
					"junctionObject": "dealContact",
					"junctionFromField": "deal_id",
					"junctionToField": "contact_id"
				}
			}
		},
		"required": ["name"]
	}`

	dealContactJunctionSchema = `{
		"type": "object",
		"properties": {
			"id": {
				"type": "string",
				"x-amp-id-field": true
			},
			"deal_id": {
				"type": "string"
			},
			"contact_id": {
				"type": "string"
			}
		},
		"required": ["deal_id", "contact_id"]
	}`
)

// setupAssociationConnector creates a connector with schemas that have associations
func setupAssociationConnector(t *testing.T) *Connector {
	t.Helper()

	rawSchemas := map[string][]byte{
		"account":            []byte(accountSchemaWithRaw),
		"contact":            []byte(contactSchemaWithForeignKey),
		"opportunity":        []byte(opportunitySchemaWithReverseLookup),
		"opportunityContact": []byte(opportunityContactSchema),
		"deal":               []byte(dealSchemaWithJunction),
		"dealContact":        []byte(dealContactJunctionSchema),
	}

	connector, err := NewConnector(WithRawSchemas(rawSchemas))
	require.NoError(t, err, "Failed to create connector")

	return connector
}

func TestForeignKeyAssociation_Expansion(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create an account
	accountResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"id":       "acc-1",
			"name":     "Acme Corp",
			"industry": "Technology",
		},
	})
	require.NoError(t, err)
	require.True(t, accountResult.Success)

	// Create a contact with foreign key to account
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"id":         "cont-1",
			"name":       "John Doe",
			"email":      "john@example.com",
			"account_id": "acc-1",
		},
	})
	require.NoError(t, err)
	require.True(t, contactResult.Success)

	// Read contact with association expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "contact",
		Fields:            datautils.NewStringSet("id", "name", "email", "account_id"),
		AssociatedObjects: []string{"account_id"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	// Verify association was expanded
	contact := readResult.Data[0]
	assert.Contains(t, contact.Associations, "account_id")
	assert.Len(t, contact.Associations["account_id"], 1)

	// Verify expanded account data
	expandedAccount := contact.Associations["account_id"][0]
	accountData := expandedAccount.Raw
	assert.Equal(t, "acc-1", accountData["id"])
	assert.Equal(t, "Acme Corp", accountData["name"])
	assert.Equal(t, "Technology", accountData["industry"])
}

func TestForeignKeyAssociation_NullForeignKey(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create a contact without a foreign key (null account_id)
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"id":    "cont-1",
			"name":  "Jane Doe",
			"email": "jane@example.com",
			// account_id is intentionally omitted
		},
	})
	require.NoError(t, err)
	require.True(t, contactResult.Success)

	// Read contact with association expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "contact",
		Fields:            datautils.NewStringSet("id", "name", "email"),
		AssociatedObjects: []string{"account_id"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	// Verify association is empty (not an error, just no related records)
	contact := readResult.Data[0]
	if associations, exists := contact.Associations["account_id"]; exists {
		assert.Empty(t, associations, "Expected no associations for null foreign key")
	}
}

func TestReverseLookupAssociation_Expansion(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create an opportunity
	opportunityResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "opportunity",
		RecordData: map[string]any{
			"id":   "opp-1",
			"name": "Big Deal",
		},
	})
	require.NoError(t, err)
	require.True(t, opportunityResult.Success)

	// Create opportunity contacts (child records)
	for i, contactID := range []string{"cont-1", "cont-2"} {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "opportunityContact",
			RecordData: map[string]any{
				"id":             "opp-cont-" + string(rune('0'+i+1)),
				"opportunity_id": "opp-1",
				"contact_id":     contactID,
			},
		})
		require.NoError(t, err)
	}

	// Read opportunity with reverse lookup expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "opportunity",
		Fields:            datautils.NewStringSet("id", "name"),
		AssociatedObjects: []string{"contacts"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	// Verify associations were expanded
	opportunity := readResult.Data[0]
	assert.Contains(t, opportunity.Associations, "contacts")
	assert.Len(t, opportunity.Associations["contacts"], 2)

	// Verify expanded contact data
	for _, assoc := range opportunity.Associations["contacts"] {
		contactData := assoc.Raw
		assert.Equal(t, "opp-1", contactData["opportunity_id"])
		assert.Contains(t, []string{"cont-1", "cont-2"}, contactData["contact_id"])
	}
}

func TestReverseLookupAssociation_NoChildren(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create an opportunity with no children
	opportunityResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "opportunity",
		RecordData: map[string]any{
			"id":   "opp-1",
			"name": "Lonely Deal",
		},
	})
	require.NoError(t, err)
	require.True(t, opportunityResult.Success)

	// Read opportunity with reverse lookup expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "opportunity",
		Fields:            datautils.NewStringSet("id", "name"),
		AssociatedObjects: []string{"contacts"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	// Verify associations are empty
	opportunity := readResult.Data[0]
	if associations, exists := opportunity.Associations["contacts"]; exists {
		assert.Empty(t, associations, "Expected no associations for record with no children")
	}
}

func TestJunctionTableAssociation_Expansion(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create a deal
	dealResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "deal",
		RecordData: map[string]any{
			"id":   "deal-1",
			"name": "Enterprise Deal",
		},
	})
	require.NoError(t, err)
	require.True(t, dealResult.Success)

	// Create contacts
	contactIDs := []string{"cont-1", "cont-2", "cont-3"}
	for _, contactID := range contactIDs {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contact",
			RecordData: map[string]any{
				"id":    contactID,
				"name":  "Contact " + contactID,
				"email": contactID + "@example.com",
			},
		})
		require.NoError(t, err)
	}

	// Create junction table records
	for i, contactID := range contactIDs {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "dealContact",
			RecordData: map[string]any{
				"id":         "dc-" + string(rune('0'+i+1)),
				"deal_id":    "deal-1",
				"contact_id": contactID,
			},
		})
		require.NoError(t, err)
	}

	// Read deal with junction table expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "deal",
		Fields:            datautils.NewStringSet("id", "name"),
		AssociatedObjects: []string{"contacts"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	// Verify associations were expanded
	deal := readResult.Data[0]
	assert.Contains(t, deal.Associations, "contacts")
	assert.Len(t, deal.Associations["contacts"], 3)

	// Verify expanded contact data
	expandedContactIDs := make([]string, 0)

	for _, assoc := range deal.Associations["contacts"] {
		contactData := assoc.Raw
		expandedContactIDs = append(expandedContactIDs, contactData["id"].(string))
	}

	assert.ElementsMatch(t, contactIDs, expandedContactIDs)
}

func TestJunctionTableAssociation_NoRelations(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create a deal with no related contacts
	dealResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "deal",
		RecordData: map[string]any{
			"id":   "deal-1",
			"name": "Solo Deal",
		},
	})
	require.NoError(t, err)
	require.True(t, dealResult.Success)

	// Read deal with junction table expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "deal",
		Fields:            datautils.NewStringSet("id", "name"),
		AssociatedObjects: []string{"contacts"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	// Verify associations are empty
	deal := readResult.Data[0]
	if associations, exists := deal.Associations["contacts"]; exists {
		assert.Empty(t, associations, "Expected no associations for deal with no related contacts")
	}
}

func TestWrite_ForeignKeyValidation_Success(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create an account
	_, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"id":   "acc-1",
			"name": "Valid Account",
		},
	})
	require.NoError(t, err)

	// Create a contact with valid foreign key (should succeed)
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"name":       "John Doe",
			"email":      "john@example.com",
			"account_id": "acc-1",
		},
	})
	require.NoError(t, err)
	assert.True(t, contactResult.Success)
}

func TestWrite_ForeignKeyValidation_InvalidReference(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Try to create a contact with invalid foreign key (account doesn't exist)
	_, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"name":       "Jane Doe",
			"email":      "jane@example.com",
			"account_id": "non-existent-account",
		},
	})

	// Should fail with invalid foreign key error
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidForeignKey)
	assert.Contains(t, err.Error(), "non-existent-account")
}

func TestWrite_ForeignKeyValidation_Update(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create accounts
	for _, id := range []string{"acc-1", "acc-2"} {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "account",
			RecordData: map[string]any{
				"id":   id,
				"name": "Account " + id,
			},
		})
		require.NoError(t, err)
	}

	// Create a contact
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"id":         "cont-1",
			"name":       "John Doe",
			"email":      "john@example.com",
			"account_id": "acc-1",
		},
	})
	require.NoError(t, err)

	// Update contact with valid foreign key (should succeed)
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordId:   contactResult.RecordId,
		RecordData: map[string]any{
			"account_id": "acc-2",
		},
	})
	require.NoError(t, err)

	// Try to update with invalid foreign key (should fail)
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordId:   contactResult.RecordId,
		RecordData: map[string]any{
			"account_id": "non-existent-account",
		},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidForeignKey)
}

func TestExpandAssociations_InvalidAssociationField(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create a contact
	_, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	})
	require.NoError(t, err)

	// Try to expand a non-existent association field
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "contact",
		Fields:            datautils.NewStringSet("id", "name", "email"),
		AssociatedObjects: []string{"non_existent_field"},
	})

	// Should not error, but the association should not be present
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	contact := readResult.Data[0]
	assert.NotContains(t, contact.Associations, "non_existent_field")
}

func TestExpandAssociations_TargetObjectNotFound(t *testing.T) {
	t.Parallel()

	// Create a schema with an association to a non-existent object
	rawSchemas := map[string][]byte{
		"contact": []byte(`{
			"type": "object",
			"properties": {
				"id": {
					"type": "string",
					"x-amp-id-field": true
				},
				"name": {
					"type": "string"
				},
				"invalid_ref": {
					"type": "string",
					"x-amp-association": {
						"associationType": "foreignKey",
						"targetObject": "nonExistentObject"
					}
				}
			},
			"required": ["name"]
		}`),
	}

	conn, err := NewConnector(WithRawSchemas(rawSchemas))
	require.NoError(t, err)

	ctx := context.Background()

	// Try to create a contact with a reference to non-existent object
	// This should fail during Write validation because the target object schema doesn't exist
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"name":        "John Doe",
			"invalid_ref": "some-id",
		},
	})

	// Validation should fail because the reference points to a non-existent record
	// (Since we can't check if the target object exists at validation time, it fails
	// when trying to validate the foreign key reference)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidForeignKey)
}

func TestExpandAssociations_MultipleAssociations(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create an account
	_, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"id":   "acc-1",
			"name": "Acme Corp",
		},
	})
	require.NoError(t, err)

	// Create a contact with account reference
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"id":         "cont-1",
			"name":       "John Doe",
			"email":      "john@example.com",
			"account_id": "acc-1",
		},
	})
	require.NoError(t, err)

	// Create a deal
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "deal",
		RecordData: map[string]any{
			"id":   "deal-1",
			"name": "Big Deal",
		},
	})
	require.NoError(t, err)

	// Create junction record
	_, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "dealContact",
		RecordData: map[string]any{
			"id":         "dc-1",
			"deal_id":    "deal-1",
			"contact_id": contactResult.RecordId,
		},
	})
	require.NoError(t, err)

	// Read deal with multiple association expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "deal",
		Fields:            datautils.NewStringSet("id", "name"),
		AssociatedObjects: []string{"contacts"},
	})
	require.NoError(t, err)
	require.Len(t, readResult.Data, 1)

	deal := readResult.Data[0]
	assert.Contains(t, deal.Associations, "contacts")
	assert.Len(t, deal.Associations["contacts"], 1)
}

func TestGetRecordsByIds_WithAssociations(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create an account
	_, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"id":   "acc-1",
			"name": "Acme Corp",
		},
	})
	require.NoError(t, err)

	// Create contacts
	contactIDs := []string{"cont-1", "cont-2"}
	for _, contactID := range contactIDs {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contact",
			RecordData: map[string]any{
				"id":         contactID,
				"name":       "Contact " + contactID,
				"email":      contactID + "@example.com",
				"account_id": "acc-1",
			},
		})
		require.NoError(t, err)
	}

	// Get records by IDs with association expansion
	records, err := conn.GetRecordsByIds(ctx, common.ReadByIdsParams{
		ObjectName:        "contact",
		RecordIds:         contactIDs,
		AssociatedObjects: []string{"account_id"},
	})
	require.NoError(t, err)
	require.Len(t, records, 2)

	// Verify associations were expanded for all records
	for _, record := range records {
		assert.Contains(t, record.Associations, "account_id")
		assert.Len(t, record.Associations["account_id"], 1)

		expandedAccount := record.Associations["account_id"][0]
		accountData := expandedAccount.Raw
		assert.Equal(t, "acc-1", accountData["id"])
		assert.Equal(t, "Acme Corp", accountData["name"])
	}
}

func TestWrite_NoValidation_WhenNoForeignKey(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create a contact without foreign key (should succeed even though account doesn't exist)
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
			// No account_id provided
		},
	})
	require.NoError(t, err)
	assert.True(t, contactResult.Success)
}

func TestExpandAssociations_Pagination(t *testing.T) {
	t.Parallel()

	conn := setupAssociationConnector(t)
	ctx := context.Background()

	// Create account
	_, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"id":   "acc-1",
			"name": "Acme Corp",
		},
	})
	require.NoError(t, err)

	// Create multiple contacts
	numContacts := 5
	for i := 0; i < numContacts; i++ {
		_, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contact",
			RecordData: map[string]any{
				"name":       "Contact " + string(rune('0'+i)),
				"email":      "contact" + string(rune('0'+i)) + "@example.com",
				"account_id": "acc-1",
			},
		})
		require.NoError(t, err)
	}

	// Read with pagination and association expansion
	readResult, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "contact",
		Fields:            datautils.NewStringSet("id", "name", "email", "account_id"),
		AssociatedObjects: []string{"account_id"},
		PageSize:          2,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), readResult.Rows)
	assert.False(t, readResult.Done)

	// Verify associations were expanded for paginated results
	for _, contact := range readResult.Data {
		assert.Contains(t, contact.Associations, "account_id")
		assert.Len(t, contact.Associations["account_id"], 1)
	}
}
