package revenuecat

import (
	"context"
	"net/http"
	"net/url"
	"time"

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
		records, err := jsonquery.New(node).ArrayOptional(recordsKey)
		if err != nil {
			return nil, err
		}

		if records == nil {
			return []*ajson.Node{}, nil
		}

		return records, nil
	}
}

func makeIncrementalFilterFunc(
	params common.ReadParams,
	nextPageFunc common.NextPageFunc,
) common.RecordsFilterFunc {
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return func(p common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
		if len(records) == 0 {
			return records, "", nil
		}

		timestampKey, ok := chooseTimestampKey(records[0])
		if !ok {
			return readhelper.MakeIdentityFilterFunc(nextPageFunc)(p, body, records)
		}

		boundary := readhelper.NewTimeBoundary()

		filtered := make([]*ajson.Node, 0, len(records))
		pageTimestamps := make([]time.Time, 0, len(records))

		for _, rec := range records {
			ts, err := extractMillisTimestamp(rec, timestampKey)
			if err != nil {
				return nil, "", err
			}

			if ts.IsZero() {
				return readhelper.MakeIdentityFilterFunc(nextPageFunc)(p, body, records)
			}

			pageTimestamps = append(pageTimestamps, ts)

			if boundary.Contains(p, ts) {
				filtered = append(filtered, rec)
			}
		}

		nextPage, err := nextPageFunc(body)
		if err != nil || nextPage == "" {
			return filtered, nextPage, err
		}

		order := inferTimeOrder(pageTimestamps)
		lastTS := pageTimestamps[len(pageTimestamps)-1]

		switch order {
		case readhelper.ReverseOrder:
			if !p.Since.IsZero() && lastTS.Before(p.Since) {
				return filtered, "", nil
			}
		case readhelper.ChronologicalOrder:
			if !p.Until.IsZero() && lastTS.After(p.Until) {
				return filtered, "", nil
			}
		default:
		}

		return filtered, nextPage, nil
	}
}

func chooseTimestampKey(record *ajson.Node) (string, bool) {
	candidates := []string{
		"updated_at",
		"last_updated_at",
		"last_seen_at",
		"created_at",
		"first_seen_at",
		"purchased_at",
	}

	for _, key := range candidates {
		val, err := jsonquery.New(record).IntegerOptional(key)
		if err == nil && val != nil {
			return key, true
		}
	}

	return "", false
}

func extractMillisTimestamp(record *ajson.Node, key string) (time.Time, error) {
	val, err := jsonquery.New(record).IntegerOptional(key)
	if err != nil {
		return time.Time{}, err
	}
	if val == nil {
		return time.Time{}, nil
	}

	return time.UnixMilli(*val), nil
}

func inferTimeOrder(timestamps []time.Time) readhelper.TimeOrder {
	if len(timestamps) < 2 {
		return readhelper.Unordered
	}

	nonDecreasing := true
	nonIncreasing := true

	for i := 1; i < len(timestamps); i++ {
		if timestamps[i].Before(timestamps[i-1]) {
			nonDecreasing = false
		}
		if timestamps[i].After(timestamps[i-1]) {
			nonIncreasing = false
		}
	}

	switch {
	case nonIncreasing && !nonDecreasing:
		return readhelper.ReverseOrder
	case nonDecreasing && !nonIncreasing:
		return readhelper.ChronologicalOrder
	case nonDecreasing && nonIncreasing:
		return readhelper.Unordered
	default:
		return readhelper.Unordered
	}
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
