package granola

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v0"
	defaultPageSize = 30 // https://docs.granola.ai/api-reference/list-notes#parameter-page-size
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	pageSize := defaultPageSize
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}
	url.WithQueryParam("page_size", strconv.Itoa(pageSize))

	if !params.Since.IsZero() {
		url.WithQueryParam("created_after", params.Since.Format(time.RFC3339))
	}

	if !params.Until.IsZero() {
		url.WithQueryParam("created_before", params.Until.Format(time.RFC3339))
	}

	if params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(ctx context.Context, params common.ReadParams,
	_ *http.Request, resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		common.MakeRecordsFunc(params.ObjectName),
		makeNextRecordsURL(),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

/*
	{
		"notes": [
		 ...
		],
		"hasMore": true,
		"cursor": "eyJjcmVkZW50aWFsfQ=="
	  }
*/
func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		cursor, err := jsonquery.New(node).StringOptional("cursor")
		if err != nil {
			return "", err
		}

		if cursor == nil {
			return "", nil
		}

		return *cursor, nil
	}
}
