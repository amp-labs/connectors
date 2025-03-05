package github

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/github/metadata"
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
		// GitHub uses Link header for pagination
		links := responseHeaders["Link"]
		if len(links) == 0 {
			return "", nil
		}

		nextLink := links[0]
		if nextLink == "" {
			return "", nil
		}

		// Extract URL from Link format
		// Format: <https://api.github.com/...>; rel="next"
		start := strings.Index(nextLink, "<")
		end := strings.Index(nextLink, ">")
		if start == -1 || end == -1 || !strings.Contains(nextLink, `rel="next"`) {
			return "", nil
		}

		return nextLink[start+1 : end], nil
	}
}
