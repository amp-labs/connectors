package breakcold

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodGet

	if getEndpointsPostMethod.Has(objectName) {
		method = http.MethodPost

		url = url.AddPath("list")
	}

	return http.NewRequestWithContext(ctx, method, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	nodepath := ""

	// The endpoint has data nodePath in the response.
	// https://developer.breakcold.com/v3/api-reference/reminders/list-reminders-with-filters-and-pagination.
	if objectName == "reminders" {
		nodepath = "data"
	}

	//  The endpoint has leads as the nodePath in the response.
	//  https://developer.breakcold.com/v3/api-reference/leads/list-leads-with-pagination-and-filters.
	if objectName == "leads" {
		nodepath = "leads"
	}

	res, err := jsonquery.New(body).ArrayOptional(nodepath)
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ArrayToMap(res)
	if err != nil {
		return nil, err
	}

	for field := range record[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}
