package lightfield

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/objectcatalog"
	"github.com/amp-labs/connectors/providers/lightfield/internal/metadata"
)

// listObjectMetadata starts from the catalog generated from the OpenAPI spec, which holds the record
// properties (id, createdAt, updatedAt...) of every object and the fixed fields of members.
// Accounts, contacts and opportunities have workspace-specific system and custom fields, which are
// described live by GET /v1/{object}/definitions.
// https://docs.lightfield.app/using-the-api/fields-and-relationships#discovering-definitions
func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	selection, err := metadata.Catalog.Select(objectNames)
	if err != nil {
		return nil, err
	}

	failures := make(map[string]error)

	for _, objectName := range datautils.NewSetFromList(objectNames).List() {
		if !objectsWithDefinitions.Has(objectName) {
			continue
		}

		definitions, err := c.fetchDefinitions(ctx, objectName)
		if err != nil {
			failures[objectName] = err

			continue
		}

		definitions.addTo(selection, objectName)
	}

	result := selection.Result()

	// Selection can only add fields, not drop an object or record an error, so an object whose
	// definitions call failed is moved from Result to Errors here instead of returning partial fields.
	for objectName, err := range failures {
		delete(result.Result, objectName)
		result.AppendError(objectName, err)
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

// Lightfield objectType of a relationship target -> object name in this connector. The API names
// its built-in objects with a `$` prefix (`$account`); the docs' examples show them bare, but the
// live responses don't.
var objectNameByType = map[string]string{ // nolint:gochecknoglobals
	"$account":     objectNameAccounts,
	"$contact":     objectNameContacts,
	"$opportunity": objectNameOpportunities,
}

func (d definitionsResponse) addTo(selection objectcatalog.Selection, objectName string) {
	for key, definition := range d.FieldDefinitions {
		valueType, values := definition.valueType()

		selection.AddField(objectName, key, common.FieldMetadata{
			DisplayName:  displayName(definition.Label, key),
			ValueType:    valueType,
			ProviderType: definition.ValueType,
			ReadOnly:     definition.ReadOnly,
			IsCustom:     isCustom(definition.System),
			Values:       values,
		})
	}

	for key, definition := range d.RelationshipDefinitions {
		// ReferenceTo stays empty for targets this connector does not read, such as `user`.
		var referenceTo []string
		if target, ok := objectNameByType[definition.ObjectType]; ok {
			referenceTo = []string{target}
		}

		selection.AddField(objectName, key, common.FieldMetadata{
			DisplayName:  displayName(definition.Label, key),
			ValueType:    relationshipValueType(definition.Cardinality),
			ProviderType: definition.Cardinality,
			IsCustom:     isCustom(definition.System),
			ReferenceTo:  referenceTo,
		})
	}
}

// relationshipValueType types a relationship by its cardinality. Read returns the relationship's
// `values` array of related record IDs, which for HAS_ONE holds the single linked record (a lookup).
// A HAS_MANY relationship is a list of IDs rather than one lookup, so it stays other.
// https://docs.lightfield.app/using-the-api/fields-and-relationships#relationship-definition-object
func relationshipValueType(cardinality string) common.ValueType {
	if cardinality == "HAS_ONE" {
		return common.ValueTypeReference
	}

	return common.ValueTypeOther
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
