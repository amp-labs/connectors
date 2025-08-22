package sellsy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/sellsy/internal/metadata"
	"github.com/spyzhov/ajson"
)

// Every request has a page limit in range [0,100].
// https://docs.sellsy.com/api/v2/#operation/get-contacts
const defaultPageSize = "100"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	readURL, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL := params.NextPage.String()

	if params.NextPage == "" {
		// This is the first, initial page for the object.
		// Page size query parameters:
		// https://docs.sellsy.com/api/v2/#operation/get-contacts
		readURL.WithQueryParam("limit", defaultPageSize)

		endpointURL = readURL.String()
	}

	method, jsonData, err := createReadOperation(readURL, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	return req, nil
}

type readSearchPayload struct {
	Filters readSearchFilters `json:"filters,omitempty"`
}

type readSearchFilters struct {
	Updated readFilterUpdate `json:"updated,omitempty"`
}

type readFilterUpdate struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// createReadOperation such that READ operation is either:
// * POST with a payload for /search based URLs.
// * GET without a payload for the rest of URLs.
func createReadOperation(
	url *urlbuilder.URL, params common.ReadParams,
) (method string, payload []byte, err error) {
	if !strings.HasSuffix(url.Path(), "/search") {
		return http.MethodGet, nil, nil
	}

	// Payload allows searching using "updated" time.
	// https://docs.sellsy.com/api/v2/#operation/search-contacts
	searchPayload := readSearchPayload{
		Filters: readSearchFilters{},
	}

	if !params.Since.IsZero() {
		searchPayload.Filters.Updated.Start = datautils.Time.FormatRFC3339inUTC(params.Since)
	}

	if !params.Until.IsZero() {
		searchPayload.Filters.Updated.End = datautils.Time.FormatRFC3339inUTC(params.Until)
	}

	payload, err = json.Marshal(searchPayload)
	if err != nil {
		return "", nil, err
	}

	return http.MethodPost, payload, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

/*
Pagination uses cursor pagination which in Sellsy documentation is referred to as "Seek" Method.
https://docs.sellsy.com/api/v2/#section/Pagination-on-list-and-search-requests

When number of records is less than the max page size this signifies that we can ignore making the next page request.

Read Response format:

	{
	  ...
	  "pagination": {
		"limit": 2,
		"count": 2,
		"total": 32,
		"offset": "WyI0Il0="
	  }
	}
*/
func makeNextRecordsURL(requestURL *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		seekOffset, err := jsonquery.New(node, "pagination").StrWithDefault("offset", "")
		if err != nil {
			return "", err
		}

		if seekOffset == "" {
			// Next page doesn't exist.
			return "", nil
		}

		counter, _ := jsonquery.New(node, "pagination").IntegerWithDefault("count", 0)
		limit, _ := jsonquery.New(node, "pagination").IntegerWithDefault("limit", 0)

		if counter < limit {
			// This is the last page.
			// The next page cannot contain more records, so stop here.
			return "", nil
		}

		nextURL, err := urlbuilder.FromRawURL(requestURL)
		if err != nil {
			return "", err
		}

		nextURL.WithQueryParam("offset", seekOffset)

		return nextURL.String(), nil
	}
}
