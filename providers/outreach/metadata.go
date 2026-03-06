package outreach

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// Schema represents the top-level JSON Hyper-Schema structure of Outreach's API spec.
// Ref: https://developers.outreach.io/api/reference/overview/
type Schema struct {
	Definitions map[string]objectDefinition `json:"definitions"`
}

type objectDefinition struct {
	Definitions map[string]fieldDefinition `json:"definitions"`
	Links       []link                     `json:"links"`
	Properties  map[string]json.RawMessage `json:"properties"`
}

type fieldDefinition struct {
	Type string `json:"type"`
	// Format is a pointer because not all fields have a format in the schema.
	// nil means the field has no format, which lets us fall back to type-based inference.
	Format   *string `json:"format"`
	ReadOnly *bool   `json:"readOnly,omitempty"`
}

type link struct {
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	Method string `json:"method"`
}

func (item dataItem) ToMapStringAny() (map[string]any, error) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DataItem: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// schema.json returns the full Schema for all standard objects as well as any
	// custom objects defined in the workspace.
	// Ref: https://developers.outreach.io/api/reference/overview/
	url, err := c.getApiURL("schema.json")
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	response, err := common.UnmarshalJSON[Schema](res)
	if err != nil {
		return nil, err
	}

	if response == nil || len(response.Definitions) == 0 {
		return nil, fmt.Errorf("%w: could not find objects schema", common.ErrMissingExpectedValues)
	}

	for _, obj := range objectNames {
		// Standard objects were originally defined with plural names when the metadata was built via
		// sampling actual records from the API, and changing them would break existing integrations.
		// We therefore accept plural names but singularize them internally to match the singular
		// keys returned by the schema API. Custom objects are kept as-is since their schema keys
		// already match the plural form.
		lookupName := obj
		if standardObjects.Has(obj) {
			lookupName = naming.NewSingularString(obj).String()
		}

		objectDefinition, ok := response.Definitions[lookupName]
		if !ok {
			metadataResult.Errors[obj] = common.ErrObjectNotSupported

			continue
		}

		objectMetadata := common.ObjectMetadata{
			Fields:      make(map[string]common.FieldMetadata),
			FieldsMap:   make(map[string]string),
			DisplayName: naming.CapitalizeFirstLetterEveryWord(obj),
		}

		metadataMapper(objectDefinition, &objectMetadata)

		metadataResult.Result[obj] = objectMetadata
	}

	return &metadataResult, nil
}

func metadataMapper(objDefinition objectDefinition, metadata *common.ObjectMetadata) {
	attributes := objDefinition.Definitions
	for field, properties := range attributes {
		metadata.AddFieldMetadata(field, common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromProviderType(properties.Type, properties.Format),
			ProviderType: properties.Type,
			ReadOnly:     properties.ReadOnly,
			Values:       nil,
		})
	}
}

func inferValueTypeFromProviderType(value string, format *string) common.ValueType {
	if format != nil {
		switch *format {
		case "date-time":
			return common.ValueTypeDateTime
		case "date":
			return common.ValueTypeDate
		default:
			// Unknown format; fall back to type-based inference below.
		}
	}

	switch value {
	case "string":
		return common.ValueTypeString
	case "number":
		return common.ValueTypeFloat
	case "boolean":
		return common.ValueTypeBoolean

	default:
		return common.ValueTypeOther
	}
}
