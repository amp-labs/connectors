package square

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "100"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	cfg, ok := objects[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, params.ObjectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, cfg.path)
	if err != nil {
		return nil, err
	}

	if cfg.supportsCursor && params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	// When the object supports it, always request the maximum page size so we
	// minimize round trips. params.PageSize is treated as an override: when the
	// caller sets it we honor that value, otherwise we fall back to the max.
	if cfg.supportsLimit {
		url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))
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

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	cfg := objects[params.ObjectName]

	return common.ParseResult(
		resp,
		common.MakeRecordsFunc(cfg.responseKey),
		makeNextRecordsURL(),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
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
