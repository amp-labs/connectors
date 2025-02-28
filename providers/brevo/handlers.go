package brevo

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
)

var (
	apiVersion = "v3"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if supportLimitAndOffset.Has(params.ObjectName) {
		if len(params.NextPage) != 0 {
			// Next page
			url.WithQueryParam("offset", params.NextPage.String())
		} else {
			// First page
			url.WithQueryParam("offset", "0")
		}

		url.WithQueryParam("limit", strconv.Itoa(pageSize))

	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(
		response,
		common.GetRecordsUnderJSONPath(responseFieldName),
		nextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}
