package teamwork

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
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

	url.WithQueryParam("pageSize", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	if !params.Since.IsZero() && objectsWithSinceQuery.Has(params.ObjectName) {
		url.WithQueryParam("updatedAfter", datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	if !params.Until.IsZero() && objectsWithUntilQuery.Has(params.ObjectName) {
		url.WithQueryParam("updatedBefore", datautils.Time.FormatRFC3339inUTC(params.Until))
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

// makeNextRecordsURL returns a NextPageFunc for Teamwork.com v3 API pagination.
//
// Extracts pageOffset from meta.page object, increments it by 1 when hasMore is
// true, and updates the page query parameter for the next request.
// See: https://apidocs.teamwork.com/guides/teamwork/how-does-paging-work#v3-endpoint-meta-data
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

		// pageOrigin is 1 because Teamwork page counting starts at page 1 (not 0).
		// pageOffset is the 0-based distance from origin in the current response.
		const pageOrigin = 1

		// CurrentPage						= pageOrigin + pageOffset
		// NextPage		= CurrentPage + 1	= pageOrigin + pageOffset + 1
		nextPageParam := pageOrigin + *pageOffset + 1
		url.WithQueryParam("page", strconv.FormatInt(nextPageParam, 10))

		return url.String(), nil
	}
}
