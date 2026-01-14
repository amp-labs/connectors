package custom

import (
	"context"
	"errors"
	"net/http"

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
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	result := &common.UpsertMetadataResult{
		Success: true,
		Fields:  make(map[string]map[string]common.FieldUpsertResult),
	}

	for objectName, fieldDefinitions := range params.Fields {
		// Each object has a list of groups to organize fields.
		groupName, err := a.getOrCreateGroupName(ctx, objectName)
		if err != nil {
			return nil, err
		}

		fields, err := a.upsertCustomFields(ctx, objectName, groupName, fieldDefinitions)
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
	ctx context.Context, objectName string, groupName string, definitions []common.FieldDefinition,
) (map[string]common.FieldUpsertResult, error) {
	fields := make(map[string]common.FieldUpsertResult)

	// Step 1: Attempt to create every field.
	// Existing fields are returned for update.
	fieldsForUpdate, err := a.createCustomFields(ctx, objectName, groupName, definitions, fields)
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
	err = a.updateCustomFields(ctx, objectName, groupName, definitionsForUpdate, fields)
	if err != nil {
		return nil, err
	}

	// Step 4: Return the complete set of results (created + updated).
	return fields, nil
}

const (
	defaultGroupNameID          = "integrationcreatedproperties"
	defaultGroupNameDisplayName = "Integration Created Properties"
)

// getOrCreateGroup returns the property group name for a custom field in HubSpot.
// If the group already exists, it is returned; otherwise, it is created automatically.
// A property group can be thought of as a folder or tag that organizes related properties.
// HubSpot requires a group when creating or updating fields, so this helper ensures one exists.
func (a *Adapter) getOrCreateGroupName(ctx context.Context, objectName string) (string, error) {
	groupName, err := a.fetchGroupName(ctx, objectName)
	if err != nil {
		return "", err
	}

	if groupName == nil {
		return a.createGroupName(ctx, objectName)
	}

	return *groupName, nil
}

// fetchGroupName retrieves the default group for the given object.
// Returns nil if the group does not exist.
func (a *Adapter) fetchGroupName(ctx context.Context, objectName string) (*string, error) {
	url, err := a.getPropertyGroupNameURL(objectName, defaultGroupNameID)
	if err != nil {
		return nil, err
	}

	response, err := a.Client.Get(ctx, url.String())
	if err != nil {
		// Check that record was not found.
		// It is not considered an error but a valid exist path.
		var httpError *common.HTTPError
		if errors.As(err, &httpError) {
			if httpError.Status == http.StatusNotFound {
				return nil, nil // nolint:nilnil
			}
		}

		return nil, err
	}

	groupNameObject, err := common.UnmarshalJSON[GroupNameModel](response)
	if err != nil {
		return nil, err
	}

	if groupNameObject == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return &groupNameObject.Name, nil
}

// createGroupName creates a new default property group for the given object.
// Returns the name of the newly created group.
func (a *Adapter) createGroupName(ctx context.Context, objectName string) (string, error) {
	url, err := a.getPropertyGroupNameCreationURL(objectName)
	if err != nil {
		return "", err
	}

	response, err := a.Client.Post(ctx, url.String(), &GroupNameModel{
		Name:         defaultGroupNameID,
		Label:        defaultGroupNameDisplayName,
		DisplayOrder: 0,
		Archived:     false,
	})
	if err != nil {
		return "", err
	}

	groupNameObject, err := common.UnmarshalJSON[GroupNameModel](response)
	if err != nil {
		return "", err
	}

	if groupNameObject == nil {
		return "", common.ErrEmptyJSONHTTPResponse
	}

	return groupNameObject.Name, nil
}

// GroupNameModel represents a HubSpot property group.
// It is used for both request payloads and API responses.
type GroupNameModel struct {
	Name         string `json:"name,omitempty"`
	Label        string `json:"label,omitempty"`
	DisplayOrder int    `json:"displayOrder,omitempty"`
	Archived     bool   `json:"archived,omitempty"`
}
