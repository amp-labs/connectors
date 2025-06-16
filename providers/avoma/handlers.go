package avoma

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, "/")
	if err != nil {
		return nil, err
	}

	if EndpointsWithResultsPath.Has(params.ObjectName) {
		url.WithQueryParam("page_size", pageSize)
		if !params.Since.IsZero() && !params.Until.IsZero() {
			url.WithQueryParam("from_date", datautils.Time.FormatRFC3339inUTC(params.Since))
			url.WithQueryParam("to_date", datautils.Time.FormatRFC3339inUTC(params.Until))
		}
	}

	if params.NextPage != "" {
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
	nodePath := ""

	if EndpointsWithResultsPath.Has(params.ObjectName) {
		nodePath = "results"
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(nodePath),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
