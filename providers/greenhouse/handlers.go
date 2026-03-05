package greenhouse

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers/greenhouse/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// maxPageSize is the maximum number of records per page allowed by the Greenhouse Harvest API.
	// https://developers.greenhouse.io/harvest.html#pagination
	maxPageSize = 500
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Cursor-based pagination: use the full URL from the Link header directly.
		return urlbuilder.New(params.NextPage.String())
	}

	// First page: build URL from scratch.
	path, err := metadata.Schemas.FindURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", path)
	if err != nil {
		return nil, err
	}

	// Respect user-provided page size, capped at the API maximum.
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	url.WithQueryParam("per_page", strconv.Itoa(pageSize))

	if !params.Since.IsZero() {
		url.WithQueryParam("updated_after", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		// Empty path means records are at the root level (response is a JSON array, not nested).
		common.ExtractOptionalRecordsFromPath(""),
		makeNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}

// Next page is communicated via `Link` header under the `next` rel.
// https://developers.greenhouse.io/harvest.html#pagination
func makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}
