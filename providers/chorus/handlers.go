package chorus

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const apiVersion = "v1"

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	res, err := jsonquery.New(body).ArrayRequired("data")
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	if len(record) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// helper to create FieldMetadata
	newField := func(name string) common.FieldMetadata {
		return common.FieldMetadata{
			DisplayName:  name,
			ValueType:    common.ValueTypeOther,
			ProviderType: "", // not available
			ReadOnly:     false,
			Values:       nil,
		}
	}

	// Attributes represent the object fields in the response. All actual data is embedded under the "attributes" field.
	// Sample response:
	// {
	//   "data": [
	//     {
	//       "attributes": {
	//         "filter_name": "string",
	//         "filter_type": "string",
	//         "field_type": "string",
	//         "filter_values": null
	//       },
	//       "type": "engagement_filter",
	//       "id": "123"
	//     }
	//   ]
	// }
	// Refer to the API response documentation at:
	// https://api-docs.chorus.ai/#f8b34d44-df36-47eb-a42e-a112aa0ec474.
	for field, value := range record[0] {
		if field == "attributes" {
			if subfields, ok := value.(map[string]any); ok {
				for subfield := range subfields {
					objectMetadata.Fields[subfield] = newField(subfield)
				}
			} else {
				return nil, common.ErrMissingFields
			}

			continue
		}

		objectMetadata.Fields[field] = newField(field)
	}

	return &objectMetadata, nil
}
