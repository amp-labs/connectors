package metadata

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

var ErrPermissionSetUpsert = errors.New("metadata: upsert PermissionSet failed")

// UpsertMetadata creates or updates the definition of custom fields in Salesforce.
//
// This method uses the Salesforce Metadata API to synchronize custom field definitions,
// and ensures that any *optional* fields (those not automatically visible to users)
// receive appropriate FieldPermissions through the Ampersand-managed Permission Set.
//
// Reference:
//
//	https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_upsertMetadata.htm
//
// Behavior summary:
//  1. Calls upsertCustomFields to create or update field definitions.
//  2. If optional fields were created, retrieves the existing field permissions.
//  3. Merges existing and new permissions into a combined permission map.
//  4. Upserts the Ampersand Permission Set to include new permissions.
//  5. Ensures the current user is assigned to that Permission Set.
//
// Optional fields in Salesforce are *not visible to any profile or user by default*.
// This function guarantees those fields are accessible after creation.
func (a *Adapter) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	// ---
	// The comments below starting with [Current state] indicate the Salesforce data state
	// and the side effects of each step at that *exit point*.
	// Rerunning UpsertMetadata will safely resume progress.
	// ---

	// Step 1: Create or update all custom fields.
	result, optionalFields, err := a.upsertCustomFields(ctx, params)
	if err != nil {
		return nil, err // [Current state]: Fields upserted, but optional ones remain invisible.
	}

	// Step 2: If there are no optional fields, no permission updates are needed.
	if len(optionalFields) == 0 {
		return result, nil
	}

	// Step 3: Fetch the existing FieldPermissions defined in the Ampersand Permission Set.
	existingFieldPermissions, err := a.fetchFieldPermissions(ctx)
	if err != nil {
		return nil, err // [Current state]: Fields upserted, but optional ones remain invisible.
	}

	// Step 4: Merge new optional field permissions with the existing set.
	// This ensures existing permissions are preserved and new fields are appended.
	combinedPermissions := datautils.MergeMaps(
		existingFieldPermissions,
		optionalFields,
	)

	// Step 5: Upsert the Ampersand Permission Set with the combined permissions.
	if err = a.upsertPermissionSet(ctx, combinedPermissions); err != nil {
		return nil, err // [Current state]: Fields upserted, but optional ones remain invisible.
	}

	// Step 6: Retrieve IDs required to assign the Permission Set to the current user.
	permissionSetID, err := a.fetchPermissionSetID(ctx)
	if err != nil {
		return nil, err // [Current state]: Fields upserted and permission set has new + old field permissions.
	}

	userID, err := a.fetchUserID(ctx)
	if err != nil {
		return nil, err // [Current state]: Fields upserted and permission set has new + old field permissions.
	}

	// Step 7: Assign the Ampersand-managed Permission Set to the current user,
	// ensuring access to all optional fields that were just created.
	if err = a.assignPermissionSetToUser(ctx, userID, permissionSetID); err != nil {
		return nil, err // [Current state]: Fields upserted and permission set has new + old field permissions.
	}

	// [Current state]: Fields upserted, permission set updated, and current user is assigned to the permission set.
	return result, nil
}

func (a *Adapter) upsertCustomFields(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, FieldPermissions, error) {
	payload, err := NewCustomFieldsPayload(params)
	if err != nil {
		return nil, nil, err
	}

	response, err := performMetadataAPICall[UpsertMetadataResponse](ctx, a, payload)
	if err != nil {
		return nil, nil, err
	}

	result, err := transformResponseToResult(response)
	if err != nil {
		return nil, nil, err
	}

	return result, payload.getOptionalFields(), nil
}

func (a *Adapter) fetchFieldPermissions(ctx context.Context) (FieldPermissions, error) {
	payload := NewReadPermissionSetPayload()

	permissionSetBody, err := performMetadataAPICall[PermissionSetResponse](ctx, a, payload)
	if err != nil {
		return nil, err
	}

	return permissionSetBody.GetFieldPermissions(), nil
}

func (a *Adapter) upsertPermissionSet(ctx context.Context, permissions FieldPermissions) error {
	payload := NewPermissionSetPayload(permissions)

	response, err := performMetadataAPICall[UpsertMetadataResponse](ctx, a, payload)
	if err != nil {
		return err
	}

	// Validate that PermissionSet was successfully created/updated.
	success := false

	for _, result := range response.Response.Results {
		if result.FullName == DefaultPermissionSetName && result.Success {
			success = true
		}
	}

	if !success {
		return fmt.Errorf("%w: failed for %s", ErrPermissionSetUpsert, DefaultPermissionSetName)
	}

	return nil
}
