package square

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 100

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	cfg, ok := objects[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, params.ObjectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, cfg.path)
	if err != nil {
		return nil, err
	}

	if cfg.readViaPOST {
		return buildPOSTRequest(ctx, cfg, params, url)
	}

	return buildGETRequest(ctx, cfg, params, url)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	cfg := objects[params.ObjectName]

	return common.ParseResult(
		resp,
		makeRecordsFunc(cfg.responseKey),
		makeNextRecordsURL(),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// makeRecordsFunc extracts the records array under responseKey. Square omits the
// key entirely on an empty last page (the body is just `{}`), so we look it up
// optionally and treat a missing key as zero records rather than an error.
func makeRecordsFunc(responseKey string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(responseKey)
	}
}

// Square paginates list endpoints with a top-level `cursor` field that is
// omitted on the last page.
//
//	{
//	  "customers": [ ... ],
//	  "cursor": "GcZjJVTwYth6PnqWQQHwx"
//	}
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

// buildPOSTRequest builds a read against a search endpoint, which takes its
// pagination and filter criteria as a JSON body rather than query params.
func buildPOSTRequest(
	ctx context.Context,
	cfg objectConfig,
	params common.ReadParams,
	url *urlbuilder.URL,
) (*http.Request, error) {
	body := map[string]any{}

	if cfg.supportsCursor && params.NextPage != "" {
		body["cursor"] = params.NextPage.String()
	}

	if cfg.supportsLimit {
		body["limit"] = pageSize(params)
	}

	if cfg.supportsTimeRange {
		if !params.Since.IsZero() {
			body["begin_time"] = params.Since.UTC().Format(time.RFC3339)
		}

		if !params.Until.IsZero() {
			body["end_time"] = params.Until.UTC().Format(time.RFC3339)
		}
	}

	// Archived items are excluded by default; read them too so archiving an item
	// doesn't silently drop it from a sync.
	if params.ObjectName == objectCatalogItems {
		body["archived_state"] = "ARCHIVED_STATE_ALL"
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func buildGETRequest(
	ctx context.Context,
	cfg objectConfig,
	params common.ReadParams,
	url *urlbuilder.URL,
) (*http.Request, error) {
	if cfg.supportsCursor && params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	if cfg.supportsLimit {
		url.WithQueryParam("limit", strconv.Itoa(pageSize(params)))
	}

	if cfg.supportsTimeRange {
		if !params.Since.IsZero() {
			url.WithQueryParam("begin_time", params.Since.UTC().Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("end_time", params.Until.UTC().Format(time.RFC3339))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// pageSize returns how many records to request per page. We always ask for the
// maximum so we minimize round trips; params.PageSize is treated as an override.
func pageSize(params common.ReadParams) int {
	if params.PageSize <= 0 {
		return defaultPageSize
	}

	return params.PageSize
}
