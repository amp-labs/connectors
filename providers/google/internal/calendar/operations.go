package calendar

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	objectNameCalendarList = "calendarList"
	objectNameEvents       = "events"

	// Page size references:
	// https://developers.google.com/workspace/calendar/api/v3/reference/calendarList/list
	// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
	defaultPageSize = 3000
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := a.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("maxResults", strconv.Itoa(defaultPageSize))

	if params.ObjectName == objectNameCalendarList {
		// This is the only object to support search by deleted items.
		// https://developers.google.com/calendar/api/v3/reference/calendarList/list
		if params.Deleted {
			url.WithQueryParam("showDeleted", "true")
		}
	}

	if params.ObjectName == objectNameEvents {
		// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
		if !params.Since.IsZero() {
			url.WithQueryParam("updatedMin", datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
		}
	}

	return url, nil
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := Schemas.LookupArrayFieldName(a.Module(), params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	// Alter current request URL to progress with the next page token.
	return func(node *ajson.Node) (string, error) {
		pageToken, err := jsonquery.New(node).StrWithDefault("nextPageToken", "")
		if err != nil {
			return "", err
		}

		if len(pageToken) == 0 {
			// Next page doesn't exist
			return "", nil
		}

		url.AddEncodingExceptions(map[string]string{
			"%3D": "=",
		})
		url.WithQueryParam("pageToken", pageToken)

		return url.String(), nil
	}
}
