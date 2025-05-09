package instantlyai

import (
	"bytes"
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

	if !directResponseEndpoints.Has(params.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
	}

	if len(params.NextPage) != 0 {
		// Next page.
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	if postEndpointsOfRead.Has(params.ObjectName) {
		return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader([]byte("{}")))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if directResponseEndpoints.Has(params.ObjectName) {
		return common.ParseResult(
			response,
			common.ExtractRecordsFromPath(""),
			makeNextRecordsURL(request.URL, params.ObjectName),
			common.GetMarshaledData,
			params.Fields,
		)
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("items"),
		makeNextRecordsURL(request.URL, params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}
