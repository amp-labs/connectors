package lightfield

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/lightfield/internal/metadata"
)

// ListObjectMetadata starts from the catalog generated from the OpenAPI spec, which holds the record
// properties (id, createdAt, updatedAt...) of every object and the fixed fields of members.
// Accounts, contacts and opportunities have workspace-specific system and custom fields, which are
// described live by GET /v1/{object}/definitions.
// https://docs.lightfield.app/using-the-api/fields-and-relationships#discovering-definitions
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	selection, err := metadata.Catalog.Select(objectNames)
	if err != nil {
		return nil, err
	}

	result := selection.Result()

	for objectName, objectMetadata := range result.Result {
		if !objectsWithDefinitions.Has(objectName) {
			continue
		}

		definitions, err := c.fetchDefinitions(ctx, objectName)
		if err != nil {
			delete(result.Result, objectName)
			result.AppendError(objectName, err)

			continue
		}

		definitions.addTo(&objectMetadata)
		result.Result[objectName] = objectMetadata
	}

	return result, nil
}

func (c *Connector) fetchDefinitions(ctx context.Context, objectName string) (*definitionsResponse, error) {
	path, err := metadata.Catalog.ReadPath(objectName)
	if err != nil {
		return nil, err
	}

	definitionsURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path, "definitions")
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Get(ctx, definitionsURL.String(), versionHeader)
	if err != nil {
		return nil, err
	}

	definitions, err := common.UnmarshalJSON[definitionsResponse](response)
	if err != nil {
		return nil, err
	}

	if definitions == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return definitions, nil
}

// https://docs.lightfield.app/api/resources/account/methods/definitions
type definitionsResponse struct {
	FieldDefinitions        map[string]fieldDefinition        `json:"fieldDefinitions"`
	RelationshipDefinitions map[string]relationshipDefinition `json:"relationshipDefinitions"`
}

type fieldDefinition struct {
	Label             string            `json:"label"`
	ValueType         string            `json:"valueType"`
	System            *bool             `json:"system"`
	ReadOnly          *bool             `json:"readOnly"`
	TypeConfiguration typeConfiguration `json:"typeConfiguration"`
}

type typeConfiguration struct {
	Options []selectOption `json:"options"`
}

type selectOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type relationshipDefinition struct {
	Label       string `json:"label"`
	Cardinality string `json:"cardinality"`
	ObjectType  string `json:"objectType"`
	System      *bool  `json:"system"`
}

// Lightfield objectType of a relationship target -> object name in this connector.
var objectNameByType = map[string]string{ // nolint:gochecknoglobals
	"account":     objectNameAccounts,
	"contact":     objectNameContacts,
	"opportunity": objectNameOpportunities,
}

func (d definitionsResponse) addTo(objectMetadata *common.ObjectMetadata) {
	for key, definition := range d.FieldDefinitions {
		valueType, values := definition.valueType()

		objectMetadata.AddFieldMetadata(key, common.FieldMetadata{
			DisplayName:  displayName(definition.Label, key),
			ValueType:    valueType,
			ProviderType: definition.ValueType,
			ReadOnly:     definition.ReadOnly,
			IsCustom:     isCustom(definition.System),
			Values:       values,
		})
	}

	// Read returns a relationship as the list of related record IDs, whatever its cardinality.
	for key, definition := range d.RelationshipDefinitions {
		var referenceTo []string
		if objectName, ok := objectNameByType[definition.ObjectType]; ok {
			referenceTo = []string{objectName}
		}

		objectMetadata.AddFieldMetadata(key, common.FieldMetadata{
			DisplayName:  displayName(definition.Label, key),
			ValueType:    common.ValueTypeOther,
			ProviderType: definition.Cardinality,
			IsCustom:     isCustom(definition.System),
			ReferenceTo:  referenceTo,
		})
	}
}

// valueType maps Lightfield value types to Ampersand ones.
// EMAIL and URL values may be a string or an array, TELEPHONE is an array,
// and ADDRESS and FULL_NAME are objects, so they are all reported as other.
// https://docs.lightfield.app/using-the-api/fields-and-relationships#field-value-types
func (f fieldDefinition) valueType() (common.ValueType, []common.FieldValue) {
	switch f.ValueType {
	case "TEXT", "MARKDOWN", "READONLY_MARKDOWN", "HTML", "SOCIAL_HANDLE":
		return common.ValueTypeString, nil
	case "NUMBER", "CURRENCY":
		return common.ValueTypeFloat, nil
	case "CHECKBOX":
		return common.ValueTypeBoolean, nil
	case "DATETIME":
		return common.ValueTypeDateTime, nil
	case "SINGLE_SELECT":
		if values := f.TypeConfiguration.values(); len(values) != 0 {
			return common.ValueTypeSingleSelect, values
		}

		return common.ValueTypeString, nil
	case "MULTI_SELECT":
		if values := f.TypeConfiguration.values(); len(values) != 0 {
			return common.ValueTypeMultiSelect, values
		}

		return common.ValueTypeOther, nil
	default:
		return common.ValueTypeOther, nil
	}
}

// Select values are option IDs (opt_..., pd_... for $pipeline); the label is for display.
func (t typeConfiguration) values() []common.FieldValue {
	values := make([]common.FieldValue, 0, len(t.Options))

	for _, option := range t.Options {
		values = append(values, common.FieldValue{
			Value:        option.ID,
			DisplayValue: option.Label,
		})
	}

	return values
}

func displayName(label, key string) string {
	if label != "" {
		return label
	}

	return key
}

func isCustom(system *bool) *bool {
	if system == nil {
		return nil
	}

	custom := !*system

	return &custom
}
