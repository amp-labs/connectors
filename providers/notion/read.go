package notion

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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
	resultsKey      = "results"
	objectUsers     = "users"
	objectUploads   = "file_uploads"
	uploadTimeField = "last_edited_time"
)

// List users has no timestamp, so Since and Until cannot be applied.
// List file uploads has no time query param. Those rows are filtered on
// last_edited_time after the response. Later pages are not stopped early
// because the list order is not documented.
// https://developers.notion.com/reference/get-users
// https://developers.notion.com/reference/list-file-uploads
var readSupportedObjects = datautils.NewStringSet( //nolint:gochecknoglobals
	objectUsers,
	objectUploads,
)

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

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	endpointURL.WithQueryParam(pageSizeKey, pageSize(params))

	return endpointURL, nil
}

func pageSize(params common.ReadParams) string {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		return strconv.Itoa(maxPageSize)
	}

	return strconv.Itoa(params.PageSize)
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	responseKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)
	if responseKey == "" {
		responseKey = resultsKey
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(responseKey),
		makeFilterFunc(params, reqURL),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func makeFilterFunc(params common.ReadParams, reqURL *urlbuilder.URL) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL(reqURL)

	if params.ObjectName != objectUploads {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		uploadTimeField,
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
