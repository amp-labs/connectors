package paddle

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		url.WithQueryParam("after", params.NextPage.String())
	}

	if supportIncrementalRead.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_at[GTE]", params.Since.Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("updated_at[LTE]", params.Until.Format(time.RFC3339))
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
		common.ExtractRecordsFromPath("data"),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
