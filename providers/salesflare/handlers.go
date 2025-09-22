package salesflare

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func (c Connector) buildSingleHandlerRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.getReadURL(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c Connector) parseSingleHandlerResponse(
	ctx context.Context, objectName string, request *http.Request, response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	fields := make(common.FieldsMetadata)

	records, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return nil, err
	}

	if records == nil || len(*records) < 1 {
		return nil, common.ErrMissingFields
	}

	for fieldName := range (*records)[0] {
		fields.AddFieldWithDisplayOnly(fieldName, naming.CapitalizeFirstLetterEveryWord(fieldName))
	}

	return common.NewObjectMetadata(
		naming.CapitalizeFirstLetterEveryWord(objectName), fields,
	), nil
}
