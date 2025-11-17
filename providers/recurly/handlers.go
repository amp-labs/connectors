package recurly

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	ApiVersionHeader = "application/vnd.recurly.v2021-02-25+json" //nolint:gochecknoglobals
	limit            = "200"                                      //nolint:gochecknoglobals
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var url *urlbuilder.URL

	var err error

	if params.NextPage != "" {
		// NextPage contains the full path with cursor,
		// e.g., "/accounts?cursor=xy0togeu9vun%3A1763384298.485542&limit=2&sort=created_at"
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL + params.NextPage.String())
		if err != nil {
			return nil, err
		}
	} else {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam("limit", limit)

		if supportIncrementalRead.Has(params.ObjectName) {
			if !params.Since.IsZero() {
				url.WithQueryParam("begin_time", params.Since.Format(time.RFC3339))
			}

			if !params.Until.IsZero() {
				url.WithQueryParam("end_time", params.Until.Format(time.RFC3339))
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", ApiVersionHeader)

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
