package okta

import (
	"context"
	"errors"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Objects that support custom fields via Schema API.
// Users and Groups have customizable profile schemas in Okta.
// Reference: https://developer.okta.com/docs/reference/api/schemas
//
//nolint:gochecknoglobals
var objectsWithCustomFields = datautils.NewStringSet(
	"users",
	"groups",
)

// schemaEndpoints maps object names to their schema API endpoints.
//
//nolint:gochecknoglobals
var schemaEndpoints = map[string]string{
	"users":  "/api/v1/meta/schemas/user/default",
	"groups": "/api/v1/meta/schemas/group/default",
}

// requestCustomFields makes an API call to get the schema describing custom fields.
// For objects that don't support custom fields, an empty map is returned.
// Returns map[fieldName]customFieldDefinition.
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]customFieldDefinition, error) {
	if !objectsWithCustomFields.Has(objectName) {
		return map[string]customFieldDefinition{}, nil
	}

	schemaPath, ok := schemaEndpoints[objectName]
	if !ok {
		return map[string]customFieldDefinition{}, nil
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, schemaPath)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	schemaResponse, err := common.UnmarshalJSON[oktaSchemaResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if schemaResponse == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	// Extract custom fields from the schema response
	fields := make(map[string]customFieldDefinition)

	if schemaResponse.Definitions.Custom.Properties != nil {
		for name, prop := range schemaResponse.Definitions.Custom.Properties {
			fields[name] = customFieldDefinition{
				Name:        name,
				Title:       prop.Title,
				Description: prop.Description,
				Type:        prop.Type,
				Required:    prop.Required,
				MinLength:   prop.MinLength,
				MaxLength:   prop.MaxLength,
				Enum:        prop.Enum,
				OneOf:       prop.OneOf,
			}
		}
	}

	return fields, nil
}

// oktaSchemaResponse represents the response from Okta Schema API.
// Reference: https://developer.okta.com/docs/reference/api/schemas
type oktaSchemaResponse struct {
	ID          string            `json:"id"`
	Schema      string            `json:"$schema"`
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Created     string            `json:"created"`
	LastUpdated string            `json:"lastUpdated"`
	Definitions schemaDefinitions `json:"definitions"`
}

type schemaDefinitions struct {
	Base   schemaSection `json:"base"`
	Custom schemaSection `json:"custom"`
}

type schemaSection struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Properties map[string]schemaProperty `json:"properties"`
	Required   []string                  `json:"required"`
}

type schemaProperty struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Type        string              `json:"type"`
	Required    bool                `json:"required"`
	MinLength   int                 `json:"minLength,omitempty"`
	MaxLength   int                 `json:"maxLength,omitempty"`
	Enum        []string            `json:"enum,omitempty"`
	OneOf       []schemaEnumOption  `json:"oneOf,omitempty"`
	Permissions []schemaPermission  `json:"permissions,omitempty"`
	Master      *schemaMaster       `json:"master,omitempty"`
	Scope       string              `json:"scope,omitempty"`
	Items       *schemaPropertyItem `json:"items,omitempty"`
}

type schemaEnumOption struct {
	Const string `json:"const"`
	Title string `json:"title"`
}

type schemaPermission struct {
	Principal string `json:"principal"`
	Action    string `json:"action"`
}

type schemaMaster struct {
	Type string `json:"type"`
}

type schemaPropertyItem struct {
	Type string `json:"type"`
}

// customFieldDefinition represents a custom field definition from Okta schema.
type customFieldDefinition struct {
	Name        string
	Title       string
	Description string
	Type        string
	Required    bool
	MinLength   int
	MaxLength   int
	Enum        []string
	OneOf       []schemaEnumOption
}

// getValueType maps Okta schema types to common.ValueType.
func (f customFieldDefinition) getValueType() common.ValueType {
	switch f.Type {
	case "string":
		if len(f.Enum) > 0 || len(f.OneOf) > 0 {
			return common.ValueTypeSingleSelect
		}

		return common.ValueTypeString
	case "integer":
		return common.ValueTypeInt
	case "number":
		return common.ValueTypeFloat
	case "boolean":
		return common.ValueTypeBoolean
	case "array":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}

// getValues returns the list of possible values for enum fields.
func (f customFieldDefinition) getValues() common.FieldValues {
	// Check OneOf first (more detailed)
	if len(f.OneOf) > 0 {
		values := make(common.FieldValues, len(f.OneOf))
		for i, option := range f.OneOf {
			values[i] = common.FieldValue{
				Value:        option.Const,
				DisplayValue: option.Title,
			}
		}

		return values
	}

	// Fall back to Enum
	if len(f.Enum) > 0 {
		values := make(common.FieldValues, len(f.Enum))
		for i, option := range f.Enum {
			values[i] = common.FieldValue{
				Value:        option,
				DisplayValue: option,
			}
		}

		return values
	}

	return nil
}

// flattenProfileFields moves custom profile fields from the nested profile object to the root level.
// This allows users to request custom fields by their name directly.
// Okta stores user/group data in a nested "profile" object.
func flattenProfileFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	profileNode, err := jsonquery.New(node).ObjectOptional("profile")
	if err != nil {
		return nil, err
	}

	if profileNode == nil {
		return root, nil
	}

	profile, err := jsonquery.Convertor.ObjectToMap(profileNode)
	if err != nil {
		return nil, err
	}

	// Move all profile fields to root level for easier field access
	maps.Copy(root, profile)

	return root, nil
}
