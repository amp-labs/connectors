package flatfile

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// nolint: gochecknoglobals
var (
	version       = "v1"
	pageSize      = "100"
	pageSizeQuery = "pageSize"
	pageQuery     = "pageNumber"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, version, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(pageSizeQuery, pageSize)
	url.WithQueryParam(pageQuery, "1") // Start with the first page

	if supportObjectSince.Has(params.ObjectName) && !params.Since.IsZero() {
		url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTC(params.Since))
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
		records(),
		nextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}
