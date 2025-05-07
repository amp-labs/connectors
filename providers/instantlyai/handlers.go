package instantlyai

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	if len(params.NextPage) != 0 {
		// Next page.
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("items"),
		makeNextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}
