package revenuecat

import (
	"context"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/revenuecat/metadata"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "v2"

	// Docs: https://www.revenuecat.com/docs/api/v2#tag/Pagination
	defaultPageSize = "100"
)

// incrementalConfig is the per-object incremental-read config:
// timestampKey is the unix-ms field representing last change; order is the API's guaranteed sort direction.
// Objects absent from the map (e.g. subscriptions) have no suitable timestamp and fall back to MakeIdentityFilterFunc.
type incrementalConfig struct {
	timestampKey string
	order        readhelper.TimeOrder
}

// Docs: https://www.revenuecat.com/docs/api/v2
// List endpoints return items newest-first (ReverseOrder) unless noted otherwise.
// Objects with no documented sort order use Unordered (safe fallback).
var objectIncrementalConfig = map[string]incrementalConfig{ //nolint:gochecknoglobals
	"customers":             {timestampKey: "last_seen_at", order: readhelper.Unordered},
	"purchases":             {timestampKey: "purchased_at", order: readhelper.ReverseOrder},
	"apps":                  {timestampKey: "created_at", order: readhelper.ReverseOrder},
	"entitlements":          {timestampKey: "created_at", order: readhelper.ReverseOrder},
	"offerings":             {timestampKey: "created_at", order: readhelper.ReverseOrder},
	"products":              {timestampKey: "created_at", order: readhelper.ReverseOrder},
	"metrics_overview":      {timestampKey: "last_updated_at", order: readhelper.Unordered},
	"integrations_webhooks": {timestampKey: "created_at", order: readhelper.ReverseOrder},
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(
		c.ProviderInfo().BaseURL,
		apiVersion,
		"projects",
		c.ProjectID,
		objectPath,
	)
	if err != nil {
		return nil, err
	}

	// List endpoints support forward pagination via `limit`.
	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	return url, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)
	nextPageFunc := nextPageFromListObject(request.URL)

	return common.ParseResultFiltered(
		params,
		resp,
		extractRecordsOptional(recordsKey),
		makeIncrementalFilterFunc(params, nextPageFunc),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

func extractRecordsOptional(recordsKey string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(recordsKey)
	}
}

// makeIncrementalFilterFunc selects the right filter for the object.
// Without time bounds it is a no-op; with bounds it uses objectIncrementalConfig to
// pick the timestamp key and ordering, falling back to identity if the object is absent.
func makeIncrementalFilterFunc(
	params common.ReadParams,
	nextPageFunc common.NextPageFunc,
) common.RecordsFilterFunc {
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	cfg, ok := objectIncrementalConfig[params.ObjectName]
	if !ok {
		// No documented timestamp field for this object; pass all records through.
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		cfg.order,
		readhelper.NewTimeBoundary(),
		cfg.timestampKey,
		readhelper.TimestampFormatUnixMs,
		nextPageFunc,
	)
}

func nextPageFromListObject(previousRequestURL *url.URL) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(root).StrWithDefault("next_page", "")
		if err != nil || nextPage == "" {
			return "", err
		}

		parsed, err := url.Parse(nextPage)
		if err != nil {
			return "", err
		}

		if parsed.Scheme == "" {
			return previousRequestURL.ResolveReference(parsed).String(), nil
		}

		return nextPage, nil
	}
}
