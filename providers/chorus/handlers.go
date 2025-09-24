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

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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

	res, err := jsonquery.New(body).ArrayOptional("data")
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
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
