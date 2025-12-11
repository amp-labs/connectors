package attio

import (
	"bytes"
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Data []map[string]any `json:"data"`
}

// This struct is used for when the response data having single data.
type objectAttribute struct {
	Data []Data `json:"data"`
}

type Data struct {
	ID struct {
		WorkspaceID string `json:"workspace_id"` //nolint:tagliatelle
		ObjectID    string `json:"object_id"`    //nolint:tagliatelle
		AttributeID string `json:"attribute_id"` //nolint:tagliatelle
	} `json:"id"`
	Title                 string    `json:"title"`
	APISlug               string    `json:"api_slug"` //nolint:tagliatelle
	Type                  string    `json:"type"`
	IsWritable            bool      `json:"is_writable"`              //nolint:tagliatelle
	IsMultiselect         bool      `json:"is_multiselect"`           //nolint:tagliatelle
	IsDefaultValueEnabled bool      `json:"is_default_value_enabled"` //nolint:tagliatelle
	CreatedAt             time.Time `json:"created_at"`               //nolint:tagliatelle
	Config                struct {
		RecordReference struct {
			AllowedObjectIDs []string `json:"allowed_object_ids"` //nolint:tagliatelle
		} `json:"record_reference"`
	} `json:"config"`
}

type objectResponse struct {
	Data struct {
		Id struct {
			WorkspaceId string `json:"workspace_id"` //nolint:tagliatelle
			ObjectId    string `json:"object_id"`    //nolint:tagliatelle
		} `json:"id"`
		ApiSlug      string    `json:"api_slug"`      //nolint:tagliatelle
		SingularNoun string    `json:"singular_noun"` //nolint:tagliatelle
		PluralNoun   string    `json:"plural_noun"`   //nolint:tagliatelle
		CreatedAt    time.Time `json:"created_at"`    //nolint:tagliatelle
	} `json:"data"`
}

// ListObjectMetadata creates metadata of object via reading objects using Attio API.
//
//nolint:funlen
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.NewListObjectMetadataResult()

	for _, obj := range objectNames {
		metadata, isCustom, err := c.getObjectAttributes(ctx, obj)
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		displayName := obj

		if isCustom {
			name, err := c.getObjectDisplayName(ctx, obj)
			if err != nil {
				metadataResult.Errors[obj] = err

				continue
			}

			if name != "" {
				displayName = name
			}
		}

		metadataResult.Result[obj] = *common.NewObjectMetadata(
			displayName, metadata,
		)
	}

	return metadataResult, nil
}

func (c *Connector) getObjectAttributes(
	ctx context.Context, obj string,
) (map[string]common.FieldMetadata, bool, error) {
	// Standard isn't a term we commonly use, but rather a concept defined by Attio itself.
	// supportAttioApi represents non-standarad/custom objects.
	isAttioStandardOrCustomObj := !supportAttioApi.Has(obj)

	var (
		url *urlbuilder.URL
		err error
	)

	if isAttioStandardOrCustomObj {
		url, err = c.getObjectAttributesURL(obj)
		if err != nil {
			return nil, false, err
		}
	} else {
		url, err = c.getApiURL(obj)
		if err != nil {
			return nil, false, err
		}
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, false, err
	}

	metadata, err := c.parseMetadataFromResponse(ctx, resp, isAttioStandardOrCustomObj)
	if err != nil {
		return nil, false, err
	}

	return metadata, isAttioStandardOrCustomObj, nil
}

// getObjectDisplayName fetches the display name for custom objects.
func (c *Connector) getObjectDisplayName(ctx context.Context, obj string) (string, error) {
	url, err := c.getObjectsURL(obj)
	if err != nil {
		return "", err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	return getDisplayName(resp)
}

func (c *Connector) parseMetadataFromResponse(ctx context.Context, resp *common.JSONHTTPResponse,
	isAttioStandardOrCustomObj bool,
) (map[string]common.FieldMetadata, error) {
	// Retrieving metadata for standard and custom objects in Attio using the api_slug field.
	if isAttioStandardOrCustomObj {
		return c.parseStandardOrCustomMetadata(ctx, resp)
	}

	return c.parseMetadata(resp)
}

// Parsing the metadata response for standard or custom objects.
func (c *Connector) parseStandardOrCustomMetadata(
	ctx context.Context, resp *common.JSONHTTPResponse,
) (map[string]common.FieldMetadata, error) {
	metadata := make(map[string]common.FieldMetadata)

	response, err := common.UnmarshalJSON[objectAttribute](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	for _, value := range response.Data {
		apiSlug := value.APISlug

		var defaultValues []common.FieldValue

		if value.Type == "select" {
			defaultValues, err = c.getDefaultValues(ctx, value)
			if err != nil {
				return nil, err
			}
		}

		if value.Type == "record-reference" || value.Type == "domain" {
			defaultValues, err = c.getOptionValuesForRecordReferenceType(ctx, value)
			if err != nil {
				return nil, err
			}
		}

		metadata[apiSlug] = common.FieldMetadata{
			DisplayName:  apiSlug,
			ValueType:    getFieldValueType(value.Type, value.IsMultiselect),
			ProviderType: value.Type,
			ReadOnly:     goutils.Pointer(!value.IsWritable),
			Values:       defaultValues,
		}
	}

	return metadata, nil
}

// Parsing the metadata response for non-standard or custom objects.
func (c *Connector) parseMetadata(resp *common.JSONHTTPResponse) (map[string]common.FieldMetadata, error) {
	metadata := make(map[string]common.FieldMetadata)

	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}
	// Using the first result data to generate the metadata.
	for k := range response.Data[0] {
		metadata[k] = common.FieldMetadata{
			DisplayName:  k,
			ValueType:    common.ValueTypeOther,
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return metadata, nil
}

// Getting display name for both standard and custom objects.
func getDisplayName(resp *common.JSONHTTPResponse) (string, error) {
	response, err := common.UnmarshalJSON[objectResponse](resp)
	if err != nil {
		return "", err
	}

	if response == nil {
		return "", common.ErrMissingExpectedValues
	}

	res := response.Data.PluralNoun
	if res == "" {
		return "", common.ErrNotFound
	}

	return res, nil
}

func getFieldValueType(field string, ismultiselect bool) common.ValueType {
	switch field {
	case "number":
		return common.ValueTypeInt
	case "text":
		return common.ValueTypeString
	case "select", "record-reference", "domain":
		if ismultiselect {
			return common.ValueTypeMultiSelect
		}

		return common.ValueTypeSingleSelect
	case "date":
		return common.ValueTypeDate
	case "timestamp":
		return common.ValueTypeDateTime
	default:
		// location, currency, interaction, actor-reference
		return common.ValueTypeOther
	}
}

func (c *Connector) getDefaultValues(ctx context.Context, o Data) (fields []common.FieldValue, err error) {
	url, err := c.getOptionsURL(o.ID.ObjectID, o.ID.AttributeID)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	objectAttribute, err := common.UnmarshalJSON[objectAttribute](resp)
	if err != nil {
		return nil, err
	}

	for _, title := range objectAttribute.Data {
		fields = append(fields, common.FieldValue{
			Value:        title.Title,
			DisplayValue: title.Title,
		})
	}

	return fields, nil
}

// getOptionValuesForRecordReferenceType retrieves the default option values for a
// record-reference (relationship) attribute type.
// For such attributes, the available options are usually linked to standard objects like
// People, Companies, Deals, Users, or Workspace.
// The function fetches each object ID listed in Config.RecordReference.AllowedObjectIDs and retrieves
// its corresponding record_id to populate the field values with both Value and DisplayValue.
func (c *Connector) getOptionValuesForRecordReferenceType(
	ctx context.Context,
	data Data,
) (fields []common.FieldValue, err error) {
	objectsIDs := data.Config.RecordReference.AllowedObjectIDs

	for _, v := range objectsIDs {
		url, err := c.getObjectReadURL(v)
		if err != nil {
			return nil, err
		}

		rsp, err := c.Client.Post(ctx, url.String(), bytes.NewReader([]byte(nil)))
		if err != nil {
			return nil, err
		}

		body, ok := rsp.Body()
		if !ok {
			return nil, common.ErrEmptyJSONHTTPResponse
		}

		values, err := jsonquery.New(body).ArrayRequired("data")
		if err != nil {
			return nil, err
		}

		if len(values) == 0 {
			return nil, nil
		}

		defaultValues := []string{}

		for _, v := range values {
			recordID, err := jsonquery.New(v, "id").StringRequired("record_id")
			if err != nil {
				return nil, err
			}

			defaultValues = append(defaultValues, recordID)
		}

		for _, identifier := range defaultValues {
			fields = append(fields, common.FieldValue{
				Value:        identifier,
				DisplayValue: identifier,
			})
		}
	}

	return fields, nil
}
