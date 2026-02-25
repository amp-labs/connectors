package workday

import (
	"context"
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
// objectsWithCustomFields lists Workday objects that support custom field definitions.
var objectsWithCustomFields = datautils.NewStringSet("workers")

// aliasRegistry maps web service alias names to their field definitions.
type aliasRegistry map[string]customFieldDefinition

// nolint:tagliatelle
type customFieldDefinition struct {
	ID              string `json:"id"`
	WebServiceAlias string `json:"webServiceAlias"`
	Descriptor      string `json:"descriptor"`
	FieldType       string `json:"fieldType"`
}

type customFieldDefinitionsResponse struct {
	Data []customFieldDefinition `json:"data"`
}

func (d customFieldDefinition) Name() string {
	return "custom_field_" + strings.ToLower(strings.ReplaceAll(d.Descriptor, " ", "_"))
}

func (d customFieldDefinition) getValueType() common.ValueType {
	switch d.FieldType {
	case "Text":
		return common.ValueTypeString
	case "Numeric":
		return common.ValueTypeFloat
	case "Boolean":
		return common.ValueTypeBoolean
	case "Date":
		return common.ValueTypeDate
	default:
		return common.ValueTypeOther
	}
}

// fetchCustomFieldDefinitions retrieves custom field definitions for the given object.
// Returns an empty registry if the object does not support custom fields.
func (c *Connector) fetchCustomFieldDefinitions(
	ctx context.Context, objectName string,
) (aliasRegistry, error) {
	if !objectsWithCustomFields.Has(objectName) {
		return aliasRegistry{}, nil
	}

	url, err := c.getCustomFieldsURL(objectName)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	definitions, err := common.UnmarshalJSON[customFieldDefinitionsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if definitions == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	registry := make(aliasRegistry)
	for _, def := range definitions.Data {
		registry[def.WebServiceAlias] = def
	}

	return registry, nil
}

func (c *Connector) getCustomFieldsURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(
		c.ProviderInfo().BaseURL, "ccx", "api", "v1", c.tenantName,
		"customObjects", objectName, "fields",
	)
}

// attachReadCustomFields returns a RecordTransformer that resolves custom field
// alias keys to human-readable names.
func (c *Connector) attachReadCustomFields(registry aliasRegistry) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		if len(registry) == 0 {
			return jsonquery.Convertor.ObjectToMap(node)
		}

		return enhanceObjectsWithCustomFieldNames(node, registry)
	}
}

func enhanceObjectsWithCustomFieldNames(
	node *ajson.Node, registry aliasRegistry,
) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	for alias, def := range registry {
		if value, ok := object[alias]; ok {
			object[def.Name()] = value
		}
	}

	return object, nil
}
