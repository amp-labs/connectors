package bentley

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if objectIncrementalSupport.Has(params.ObjectName) {
		var filters []string

		if !params.Since.IsZero() {
			filters = append(filters, "createdDateTime ge "+params.Since.Format("2006-01-02T15:04:05Z"))
		}

		if !params.Until.IsZero() {
			filters = append(filters, "createdDateTime le "+params.Until.Format("2006-01-02T15:04:05Z"))
		}

		if len(filters) > 0 {
			url.WithQueryParam("$filter", strings.Join(filters, " and "))
		}
	}

	if params.NextPage != "" {
		// Bentley returns a full URL in _links.next.href, so use it directly.
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
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
		getRecords(c.ProviderContext.Module(), params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
