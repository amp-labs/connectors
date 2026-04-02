package fastspring

import (
	"context"
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
// FastSpring documents a default of 50 for list operations (example: the "limit" query parameter on
// List all accounts). https://developer.fastspring.com/reference/list-all-accounts
const (
	defaultPageSize  = "50"
	defaultEventDays = "30" // max 30 per API; used when requiresDaysParam is true
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

	if requiresDaysParam(params.ObjectName) {
		url.WithQueryParam("days", defaultEventDays)
		// Incremental range for Events list APIs: optional begin/end (YYYY-MM-DD) alongside required days.
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

// requiresDaysParam reports whether the list request must include a "days" query parameter.
// Processed events: https://developer.fastspring.com/reference/list-all-processed-events
// Unprocessed events: https://developer.fastspring.com/reference/list-all-unprocessed-events
func requiresDaysParam(objectName string) bool {
	switch objectName {
	case "events-processed", "events-unprocessed":
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
		common.GetMarshaledData,
		params.Fields,
	)
}

// recordsForRead extracts list rows: string IDs are wrapped to one-field maps by
// object type; objects are left as-is. Uses ArrayOptional because FastSpring may
// omit the array key when empty.
func recordsForRead(objectName, recordsKey string) common.RecordsFunc {
	idField := stringIDFieldForListObject(objectName)

	return func(node *ajson.Node) ([]map[string]any, error) {
		// FastSpring often omits the array key when there are no rows (e.g. empty catalog);
		// ArrayRequired would fail with ErrKeyNotFound.
		arr, err := jsonquery.New(node).ArrayOptional(recordsKey)
		if err != nil {
			return nil, err
		}

		if len(arr) == 0 {
			return []map[string]any{}, nil
		}

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
}

func stringIDFieldForListObject(objectName string) string {
	switch objectName {
	case "accounts":
		return "id"
	case "orders":
		return "order"
	case "products":
		return "path"
	case "subscriptions":
		return "subscription"
	default:
		return "id"
	}
}

// nextPageFromIntegerCounter builds a NextPageFunc that reads a numeric "nextPage"
// field from the response root and maps it to the "page" query parameter.
func nextPageFromIntegerCounter(previousRequestURL *url.URL) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		if previousRequestURL == nil {
			return "", nil
		}

		nextPage, err := jsonquery.New(root).IntegerWithDefault("nextPage", 0)
		if err != nil || nextPage == 0 {
			return "", err
		}

		cloned := *previousRequestURL
		q := cloned.Query()
		q.Set("page", strconv.FormatInt(nextPage, 10))
		cloned.RawQuery = q.Encode()

		return cloned.String(), nil
	}
}
