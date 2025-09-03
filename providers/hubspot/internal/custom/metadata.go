package custom

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// UpsertMetadata creates or updates the definition of a custom field.
//
// This operation manages the field schema in HubSpot via API. Note that while
// field definitions can be created and updated programmatically, property
// validation rules (such as regex, ranges, or character limits) can only be
// configured manually in the HubSpot dashboard and are not exposed through the API.
//
// See: https://developers.hubspot.com/docs/api-reference/crm-property-validations-v3/guide
func (a *Adapter) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	result := &common.UpsertMetadataResult{
		Success: true,
		Fields:  make(map[string]map[string]common.FieldUpsertResult),
	}

	for objectName, fieldDefinitions := range params.Fields {
		fields, err := a.upsertCustomFields(ctx, objectName, fieldDefinitions)
		if err != nil {
			return nil, err
		}

		result.Fields[objectName] = fields
	}

	return result, nil
}

// upsertCustomFields ensures that all given field definitions exist by
// performing an upsert operation against HubSpot.
//
// The algorithm is:
//  1. Create all fields.
//  2. If creation fails because a field already exists, mark it for update.
//  3. Update the marked fields.
//  4. Return a map of results for all created and updated fields.
func (a *Adapter) upsertCustomFields(
	ctx context.Context, objectName string, definitions []common.FieldDefinition,
) (map[string]common.FieldUpsertResult, error) {
	fields := make(map[string]common.FieldUpsertResult)

	// Step 1: Attempt to create every field.
	// Existing fields are returned for update.
	fieldsForUpdate, err := a.createCustomFields(ctx, objectName, definitions, fields)
	if err != nil {
		return nil, err
	}

	// Step 2: Collect definitions for fields that need updating.
	fieldDefinitionsMap := datautils.SliceToMap(definitions, func(value common.FieldDefinition) string {
		return value.FieldName
	})
	definitionsForUpdate, _ := fieldDefinitionsMap.Select(fieldsForUpdate)

	// Step 3: Update the fields that already exist.
	// Any failed response here cannot be resolved and therefore will be surfaced.
	err = a.updateCustomFields(ctx, objectName, definitionsForUpdate, fields)
	if err != nil {
		return nil, err
	}

	// Step 4: Return the complete set of results (created + updated).
	return fields, nil
}
