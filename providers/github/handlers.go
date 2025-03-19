package github

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers/github/metadata"
	"github.com/spyzhov/ajson"
)

const (
	pageSizeKey = "per_page"
	pageSize    = "100"
	pageKey     = "page"
	pageNumber  = "1"
	sinceKey    = "since"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if supportPagination.Has(params.ObjectName) {
		url.WithQueryParam(pageSizeKey, pageSize)
		url.WithQueryParam(pageKey, pageNumber)
	}

	if supportSince.Has(params.ObjectName) && !params.Since.IsZero() {
		// https://docs.github.com/en/rest/gists/gists?apiVersion=2022-11-28#list-gists-for-the-authenticated-user
		url.WithQueryParam(sinceKey, datautils.Time.FormatRFC3339inUTC(params.Since))
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
		getRecords(params.ObjectName, c.Module()),
		makeNextRecordsURL(response.Headers),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(responseHeaders http.Header) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextURL := httpkit.HeaderLink(&common.JSONHTTPResponse{Headers: responseHeaders}, "next")
		if nextURL == "" {
			return "", nil
		}
		return nextURL, nil
	}
}
