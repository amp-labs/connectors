package pardot

import (
	"context"
	_ "embed"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	fileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas) // nolint:gochecknoglobals

	// Schemas is cached Object schemas.
	Schemas = pardotSchemas{ // nolint:gochecknoglobals
		Metadata: fileManager.MustLoadSchemas(),
	}
)

func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := Schemas.Select(objectNames)
	if err != nil {
		return nil, err
	}

	for objectName, metadata := range result.Result {
		// Prospects is the only object that supports custom fields.
		// A dedicated API call to collect "prospect custom fields" will happen.
		if objectName == "prospects" {
			if err = a.fetchProspectsCustomFields(ctx, &metadata); err != nil {
				return nil, err
			}

			result.Result[objectName] = metadata
		}
	}

	return result, nil
}

type pardotSchemas struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, any]
}

func (s *pardotSchemas) Select(objectNames []string) (*common.ListObjectMetadataResult, error) {
	// Case-insensitive object names.
	objects := make([]string, len(objectNames))
	for index, name := range objectNames {
		objects[index] = strings.ToLower(name)
	}

	return s.Metadata.Select(providers.ModuleSalesforceAccountEngagement, objects)
}

func (a *Adapter) fetchProspectsCustomFields(
	ctx context.Context, metadata *common.ObjectMetadata,
) error {
	// https://developer.salesforce.com/docs/marketing/pardot/guide/custom-field-v5.html#custom-field-query
	url, err := a.getURL("custom-fields")
	if err != nil {
		return err
	}

	url.WithQueryParam("fields", "id,name,fieldId,type,isRequired")

	endpoint := new(url.String())

	for endpoint != nil {
		resp, err := a.JSONHTTPClient().Get(ctx, *endpoint, common.Header{
			Key:   "Pardot-Business-Unit-Id",
			Value: a.businessUnitID,
		})
		if err != nil {
			return err
		}

		response, err := common.UnmarshalJSON[prospectCustomFieldsResponse](resp)
		if err != nil {
			return err
		}

		for _, field := range response.Values {
			if field.IsArrayField() {
				// Array fields should not be advertised by ListObjectMetadata.
				// Underlying endpoint used by Read(Prospects) fails when such fields are used.
				// Individual, per-record request would work, but would be deemed costly.
				continue
			}

			metadata.AddFieldMetadata(field.FieldName(), common.FieldMetadata{
				DisplayName:  field.DisplayName,
				ValueType:    field.ValueType(),
				ProviderType: field.Type,
				ReadOnly:     new(false), // can write, modify data
				IsCustom:     new(true),
				IsRequired:   field.IsRequired,
				Values:       nil, // API does not return anything even for radio buttons or dropdowns.
				ReferenceTo:  nil, // not applicable
			})
		}

		// Go to the next page if it exists.
		endpoint = response.NextPageURL
	}

	// Side effects applied to `metadata`.
	return nil
}

type prospectCustomFieldsResponse struct {
	NextPageURL *string                             `json:"nextPageUrl"`
	Values      []prospectCustomFieldsResponseValue `json:"values"`
}

type prospectCustomFieldsResponseValue struct {
	ID          int    `json:"id"`
	FieldID     string `json:"fieldId"`
	DisplayName string `json:"name"`
	Type        string `json:"type"`
	IsRequired  *bool  `json:"isRequired"`
}

// FieldName returns a requestable field by Read operation.
// Custom field id must be suffixed with `__c` to indicate that it is custom field.
// This is a common practice in Salesforce.
func (v prospectCustomFieldsResponseValue) FieldName() string {
	return v.FieldID + "__c"
}

// ValueType returns a mapping of a field type to Ampersand defined field types.
// https://developer.salesforce.com/docs/marketing/pardot/guide/custom-field-v5.html#required-editable-fields
func (v prospectCustomFieldsResponseValue) ValueType() common.ValueType {
	switch strings.ToLower(v.Type) {
	case "text", "textarea", "radio button", "dropdown", "hidden", "crm user":
		return common.ValueTypeString
	case "multi-select", "checkbox":
		return common.ValueTypeMultiSelect
	case "number":
		return common.ValueTypeFloat
	case "date":
		// Salesforce performs validation on a string. It must be properly formatted date.
		return common.ValueTypeDate
	default:
		return common.ValueTypeOther
	}
}

// IsArrayField reports if the custom field of a prospect holds an array.
// Alternatively it can be a string, a float.
//
// Array custom fields are not supported by Prospect Query endpoint.
// https://developer.salesforce.com/docs/marketing/pardot/guide/prospect-v5.html#requesting-custom-fields
func (v prospectCustomFieldsResponseValue) IsArrayField() bool {
	switch strings.ToLower(v.Type) {
	case "multi-select":
		return true
	case "checkbox":
		return true
	default:
		return false
	}
}
