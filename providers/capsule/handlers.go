package capsule

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
	"github.com/spyzhov/ajson"
)

// DefaultPageSize
// https://developer.capsulecrm.com/v2/overview/reading-from-the-api
const DefaultPageSize = "100"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("perPage", DefaultPageSize)

	if !params.Since.IsZero() {
		url.WithQueryParam("since", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	if len(params.AssociatedObjects) != 0 {
		url.WithQueryParam("embed", strings.Join(params.AssociatedObjects, ","))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}

// Next page is communicated via `Link` header under the `next` rel.
// https://developer.capsulecrm.com/v2/overview/reading-from-the-api
func makeNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}
