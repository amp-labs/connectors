package notion

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/notion/metadata"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v1"
	notionVersion   = "2026-03-11"
	maxPageSize     = 100
	pageSizeKey     = "page_size"
	startCursorKey  = "start_cursor"
	uploadTimeField = "last_edited_time"
)

// Read is limited to workspace-scoped GET lists with no required parent id or
// scope query param. Users and file uploads match that model.
//
// List users has no timestamp, so Since and Until cannot be applied.
// List file uploads has no time query param. Those rows are filtered on
// last_edited_time after the response. Later pages are not stopped early
// because the list order is not documented.
// https://developers.notion.com/reference/get-users
// https://developers.notion.com/reference/list-file-uploads
//
// Other Notion list endpoints:
//   - custom_emojis: same pagination shape (`results`); parent-free. Not in Read
//     yet; easy to add to metadata and readRegistry when product wants it.
//   - views: GET list but requires database_id or data_source_id; not parent-free.
//     https://developers.notion.com/reference/list-views
//
// Databases, data sources, and pages are central to Notion but have no
// parent-free list for Read. POST /v1/search can discover them without a time
// window, so it is a poor fit for incremental Read. Subscribe (webhooks) may
// cover those objects even when Read cannot, depending on Notion's event surface.

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !readSupportedObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	endpointURL, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Notion-Version", notionVersion)

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	endpointURL.WithQueryParam(pageSizeKey, readhelper.PageSizeWithDefaultStr(params, strconv.Itoa(maxPageSize)))

	return endpointURL, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	spec, err := readEndpointFor(params.ObjectName)
	if err != nil {
		return nil, err
	}

	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(spec.responseKey),
		makeFilterFunc(spec, reqURL),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func makeFilterFunc(spec readEndpoint, reqURL *urlbuilder.URL) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL(reqURL)

	if spec.timeField == "" {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		spec.timeField,
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

		reqLink.WithQueryParam(startCursorKey, nextCursor)

		return reqLink.String(), nil
	}
}
