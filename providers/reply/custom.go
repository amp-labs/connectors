package reply

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// objectCustomFieldsOwner is the only object that carries user-defined custom
// fields. Contact records expose them as customFields: [{key, value}], where
// key is the numeric definition id serialized as a string.
const objectCustomFieldsOwner = "contacts"

// customFieldDefinition is one entry of the GET /v3/custom-fields response.
// https://docs.reply.io/api-reference/custom-fields/list-all-custom-fields
type customFieldDefinition struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	FieldType string `json:"fieldType"`
	Metadata  string `json:"metadata"`
	OrgWide   bool   `json:"orgWide"`
}

func (d customFieldDefinition) valueType() common.ValueType {
	switch d.FieldType {
	case "number":
		return common.ValueTypeFloat
	case "text":
		return common.ValueTypeString
	default:
		return common.ValueTypeOther
	}
}

// fetchCustomFieldDefinitions returns every custom field definition. The
// endpoint responds with a root-level array and is not paginated.
func (c *Connector) fetchCustomFieldDefinitions(ctx context.Context) ([]customFieldDefinition, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "custom-fields")
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	definitions, err := common.UnmarshalJSON[[]customFieldDefinition](response)
	if err != nil {
		return nil, err
	}

	if definitions == nil {
		return []customFieldDefinition{}, nil
	}

	return *definitions, nil
}
