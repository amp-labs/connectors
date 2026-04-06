package fastspring

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/fastspring/metadata"
	"github.com/spyzhov/ajson"
)

// List endpoints use limit + page; when ReadParams.PageSize is unset we send limit=defaultPageSize.
// FastSpring documents a default of 50 for list operations (e.g. "limit" on List all accounts) and does
// not document an upper bound. https://developer.fastspring.com/reference/list-all-accounts
// We use 1000 here as an arbitrary larger page size to reduce round trips;
// callers can override via ReadParams.PageSize.
const (
	defaultPageSize  = "1000"
	defaultEventDays = "30" // max 30 per API; used for event list reads (isEventObject)
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	if params.ObjectName == "" {
		return nil, common.ErrMissingObjects
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if isEventObject(params.ObjectName) {
		url.WithQueryParam("days", defaultEventDays)
		// Event list APIs: required "days" plus optional begin/end (YYYY-MM-DD) from ReadParams.Since/Until.
		// https://developer.fastspring.com/reference/events
		if !params.Since.IsZero() {
			url.WithQueryParam("begin", params.Since.Format(time.DateOnly))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("end", params.Until.Format(time.DateOnly))
		}
	}

	// FastSpring list endpoints support basic cursor-like pagination via limit + page.
	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	// When no explicit page is provided via NextPage, we start from page=1.
	url.WithQueryParam("page", "1")

	return url, nil
}

// isEventObject reports whether the object is a processed/unprocessed events list, which uses the
// event query parameters: "days" (required), and optionally "begin" / "end" for incremental range.
// Processed: https://developer.fastspring.com/reference/list-all-processed-events
// Unprocessed: https://developer.fastspring.com/reference/list-all-unprocessed-events
func isEventObject(objectName string) bool {
	switch objectName {
	case objectEventsProcessed, objectEventsUnprocessed:
		return true
	default:
		return false
	}
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	records := recordsForRead(params.ObjectName, recordsKey)

	return common.ParseResult(
		resp,
		records,
		nextPageFromIntegerCounter(request.URL),
		readhelper.MakeGetMarshaledDataWithId(stringIDFieldForListObject(params.ObjectName)),
		params.Fields,
	)
}

// recordsForRead normalizes the list under the schema responseKey (e.g. "accounts", "events").
// FastSpring returns either JSON objects or JSON string IDs depending on the endpoint; examples:
//
//	{ "accounts": [ {"id":"x","account":"y"} ] }
//	{ "accounts": [ "id1", "id2" ] }
//
// Objects are returned as-is in Raw; string elements become { "<idField>": "<id>" }
// using the IdFieldQuery.Field from stringIDFieldForListObject (flat root keys only).
// The response key may be absent when empty — jsonquery.ArrayOptional yields an empty slice without error.
// If the value is a single string instead of an array, we treat it as one row (fallback after ErrNotArray).
func recordsForRead(objectName, recordsKey string) common.RecordsFunc {
	idField := stringIDFieldForListObject(objectName).Field

	return func(node *ajson.Node) ([]map[string]any, error) {
		return recordsFromArrayOrStringKey(node, recordsKey, idField)
	}
}

func recordsFromArrayOrStringKey(node *ajson.Node, recordsKey, idField string) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).ArrayOptional(recordsKey)
	if err != nil {
		if errors.Is(err, jsonquery.ErrNotArray) {
			return recordsFromSingleStringRow(node, recordsKey, idField, err)
		}

		return nil, err
	}

	if len(arr) == 0 {
		return []map[string]any{}, nil
	}

	return mapRecordNodesToMaps(arr, idField)
}

func recordsFromSingleStringRow(
	node *ajson.Node,
	recordsKey, idField string,
	arrayErr error,
) ([]map[string]any, error) {
	str, serr := jsonquery.New(node).StringOptional(recordsKey)
	if serr != nil {
		return nil, arrayErr
	}

	if str != nil && *str != "" {
		return []map[string]any{map[string]any{idField: *str}}, nil
	}

	return nil, arrayErr
}

func mapRecordNodesToMaps(arr []*ajson.Node, idField string) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(arr))

	for _, v := range arr {
		switch {
		case v.IsObject():
			m, convErr := jsonquery.Convertor.ObjectToMap(v)
			if convErr != nil {
				return nil, convErr
			}

			out = append(out, m)
		case v.IsString():
			s, strErr := v.GetString()
			if strErr != nil {
				return nil, strErr
			}

			out = append(out, map[string]any{idField: s})
		default:
			return nil, jsonquery.ErrNotObject
		}
	}

	return out, nil
}

func stringIDFieldForListObject(objectName string) readhelper.IdFieldQuery {
	switch objectName {
	case objectAccounts:
		return readhelper.NewIdField("id")
	case objectOrders:
		return readhelper.NewIdField("order")
	case objectProducts:
		return readhelper.NewIdField("path")
	case objectSubscriptions:
		return readhelper.NewIdField("subscription")
	default:
		return readhelper.NewIdField("id")
	}
}

// nextPageFromIntegerCounter builds a NextPageFunc that reads a numeric "nextPage"
// field from the response root and maps it to the "page" query parameter on the prior request URL.
func nextPageFromIntegerCounter(previousRequestURL *url.URL) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		if previousRequestURL == nil {
			return "", nil
		}

		nextPage, err := jsonquery.New(root).IntegerWithDefault("nextPage", 0)
		if err != nil || nextPage == 0 {
			return "", err
		}

		// Re-parse from string so we do not mutate the live request URL when the builder
		// serializes (it sets RawQuery on the delegate).
		cloned, err := url.Parse(previousRequestURL.String())
		if err != nil {
			return "", err
		}

		u, err := urlbuilder.FromRawURL(cloned)
		if err != nil {
			return "", err
		}

		u.WithQueryParam("page", strconv.FormatInt(nextPage, 10))

		return u.String(), nil
	}
}
