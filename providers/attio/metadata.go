package attio

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Data []map[string]any `json:"data"`
}

// This struct is used for when the response data having single data.
type objectAttribute struct {
	Data []struct {
		ID struct {
			WorkspaceID string `json:"workspace_id"` //nolint:tagliatelle
			ObjectID    string `json:"object_id"`    //nolint:tagliatelle
			AttributeID string `json:"attribute_id"` //nolint:tagliatelle
		} `json:"id"`
		Title                 string  `json:"title"`
		Description           *string `json:"description"`
		APISlug               string  `json:"api_slug"` //nolint:tagliatelle
		Type                  string  `json:"type"`
		IsSystemAttribute     bool    `json:"is_system_attribute"`      //nolint:tagliatelle
		IsWritable            bool    `json:"is_writable"`              //nolint:tagliatelle
		IsRequired            bool    `json:"is_required"`              //nolint:tagliatelle
		IsUnique              bool    `json:"is_unique"`                //nolint:tagliatelle
		IsMultiselect         bool    `json:"is_multiselect"`           //nolint:tagliatelle
		IsDefaultValueEnabled bool    `json:"is_default_value_enabled"` //nolint:tagliatelle
		IsArchived            bool    `json:"is_archived"`              //nolint:tagliatelle
		DefaultValue          struct {
			Type     string `json:"type"`
			Template string `json:"template"`
		} `json:"default_value"` //nolint:tagliatelle
		Relationship struct {
			ID struct {
				WorkspaceID string `json:"workspace_id"` //nolint:tagliatelle
				ObjectID    string `json:"object_id"`    //nolint:tagliatelle
				AttributeID string `json:"attribute_id"` //nolint:tagliatelle
			} `json:"id"`
		} `json:"relationship"`
		CreatedAt time.Time `json:"created_at"` //nolint:tagliatelle
		Config    struct {
			Currency struct {
				DefaultCurrencyCode *string `json:"default_currency_code"` //nolint:tagliatelle
				DisplayType         *string `json:"display_type"`          //nolint:tagliatelle
			} `json:"currency"`
			RecordReference struct {
				AllowedObjectIDs []string `json:"allowed_object_ids"` //nolint:tagliatelle
			} `json:"record_reference"` //nolint:tagliatelle
		} `json:"config"`
	} `json:"data"`
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
	isAttioStandardOrCustomObj := !supportAttioApiObj.Has(obj)

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

	metadata, err := parseMetadataFromResponse(resp, isAttioStandardOrCustomObj)
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

func parseMetadataFromResponse(resp *common.JSONHTTPResponse,
	isAttioStandardOrCustomObj bool,
) (map[string]common.FieldMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	metadata := make(map[string]common.FieldMetadata)

	// Retrieving metadata for standard and custom objects in Attio using the api_slug field.
	if isAttioStandardOrCustomObj {
		response, err := common.UnmarshalJSON[objectAttribute](resp)
		if err != nil {
			return nil, err
		}

		for _, value := range response.Data {
			apiSlug := value.APISlug
			metadata[apiSlug] = common.FieldMetadata{
				DisplayName:  apiSlug,
				ValueType:    getFieldValueType(value.Type),
				ProviderType: value.Type,
				ReadOnly:     false,
				Values:       nil,
			}
		}
	} else {
		// Using the first result data to generate the metadata.
		for k := range response.Data[0] {
			metadata[k] = common.FieldMetadata{
				DisplayName:  k,
				ValueType:    common.ValueTypeOther,
				ProviderType: "", // not available
				ReadOnly:     false,
				Values:       nil,
			}
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

func getFieldValueType(field string) common.ValueType {
	switch field {
	case "number":
		return common.ValueTypeInt
	case "text":
		return common.ValueTypeString
	case "select":
		return common.ValueTypeSingleSelect
	case "date":
		return common.ValueTypeDate
	case "timestamp":
		return common.ValueTypeDateTime
	default:
		// location, currency, interaction, actor-reference, record-reference
		return common.ValueTypeOther
	}
}
