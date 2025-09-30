package teamwork

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/teamwork/internal/metadata"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "500"

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

	url.WithQueryParam("pageSize", defaultPageSize)

	if !params.Since.IsZero() {
		url.WithQueryParam("updatedAfter", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pageObject := jsonquery.New(node, "meta", "page")

		hasMore, err := pageObject.BoolWithDefault("hasMore", false)
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil
		}

		pageOffset, err := pageObject.IntegerOptional("pageOffset")
		if err != nil {
			return "", err
		}

		if pageOffset == nil {
			return "", nil
		}

		nextPageOffset := *pageOffset + 1
		url.WithQueryParam("page", strconv.FormatInt(nextPageOffset, 10))

		return url.String(), nil
	}
}
