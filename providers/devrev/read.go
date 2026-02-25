package devrev

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/devrev/metadata"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "100" // doc default 50; max 100 (from testing)

// buildReadRequest builds the HTTP request for listing objects.
// Pagination is cursor-based (next_cursor). For objects in objectsWithModifiedDateFilter,Since/Until
// are sent as modified_date.after and modified_date.before; other objects
// are filtered by modified_date on the client.
// See https://devrev.dev/docs.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		// Next page
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}
	// First page
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)
	url.WithQueryParam("limit", pageSize)
	// if object supports modified date filter, add the since and until query params
	if objectsWithModifiedDateFilter.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("modified_date.after", params.Since.UTC().Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("modified_date.before", params.Until.UTC().Format(time.RFC3339))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}
	// Get the response key
	/*{
		"accounts": [
		  ....
		],
		"next_cursor": "cursor_page_2",
		"total": 2
	  }
	*/
	responseKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)
	if responseKey == "" {
		responseKey = params.ObjectName
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(responseKey),
		makeFilterFunc(params, reqURL),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

// if object supports modified date filter, return identity filter
// otherwise return time filter.
func makeFilterFunc(params common.ReadParams, reqURL *urlbuilder.URL) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL(reqURL)

	if objectsWithModifiedDateFilter.Has(params.ObjectName) {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		"modified_date",
		time.RFC3339,
		nextPageFunc,
	)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextCursor, err := jsonquery.New(node).TextWithDefault("next_cursor", "")
		if err != nil {
			return "", err
		}

		if nextCursor == "" {
			return "", nil
		}

		reqLink.WithQueryParam("cursor", nextCursor)

		return reqLink.String(), nil
	}
}
