package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/google/internal/core"
	"github.com/spyzhov/ajson"
)

const (
	objectNameCalendarList = "calendarList"
	objectNameEvents       = "events"
	objectNameSettings     = "settings"
	objectNameACL          = "acl"

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

	switch params.ObjectName {
	case objectNameEvents:
		applyEventsQueryParams(url, params)
	case objectNameCalendarList:
		url.WithQueryParam("maxResults", strconv.Itoa(defaultPageSize))

		// This is the only object to support search by deleted items.
		// https://developers.google.com/calendar/api/v3/reference/calendarList/list
		if params.Deleted {
			url.WithQueryParam("showDeleted", "true")
		}
	default:
		url.WithQueryParam("maxResults", strconv.Itoa(defaultPageSize))
	}

	return url, nil
}

// applyEventsQueryParams sets the events.list query params shared by the primary-calendar
// (buildReadURL) and all-calendars (buildEventsURLForCalendar) read paths.
//
// Defaults preserve the historical behavior: the default page size plus updatedMin derived
// from ReadParams.Since. Additional tuning (time window, single-event expansion, deleted
// events, ordering, event-type filter) is opt-in via the Google ReadParamsOpts carried on
// ReadParams.Opts; an empty or mismatched Opts leaves only the defaults in place.
//
// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
func applyEventsQueryParams(url *urlbuilder.URL, params common.ReadParams) {
	opts, _ := params.Opts.(core.ReadParamsOpts)

	maxResults := defaultPageSize
	if opts.MaxResults > 0 {
		maxResults = opts.MaxResults
	}

	url.WithQueryParam("maxResults", strconv.Itoa(maxResults))

	if !params.Since.IsZero() {
		url.WithQueryParam("updatedMin", datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
	}

	if !opts.TimeMin.IsZero() {
		url.WithQueryParam("timeMin", datautils.Time.FormatRFC3339inUTCWithMilliseconds(opts.TimeMin))
	}

	if !opts.TimeMax.IsZero() {
		url.WithQueryParam("timeMax", datautils.Time.FormatRFC3339inUTCWithMilliseconds(opts.TimeMax))
	}

	if opts.SingleEvents {
		url.WithQueryParam("singleEvents", "true")
	}

	if opts.ShowDeleted {
		url.WithQueryParam("showDeleted", "true")
	}

	if opts.OrderBy != "" {
		url.WithQueryParam("orderBy", opts.OrderBy)
	}

	if len(opts.EventTypes) > 0 {
		url.WithQueryParamList("eventTypes", opts.EventTypes)
	}
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

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := a.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if len(params.RecordId) != 0 {
		method = http.MethodPatch

		url.AddPath(params.RecordId)
	}

	if params.ObjectName == objectNameCalendarList {
		if needsCalendarColorQueryParam(params) {
			url.WithQueryParam("colorRgbFormat", "true")
		}
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

// When either background or foreground is specified this means
// we must attach a query parameter for the Write request to succeed.
//
// https://developers.google.com/calendar/api/v3/reference/calendarList/insert
func needsCalendarColorQueryParam(params common.WriteParams) bool {
	properties, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return false
	}

	triggerFields := []string{
		"foregroundColor", "backgroundColor",
	}

	for _, field := range triggerFields {
		if _, ok := properties[field]; ok {
			return true
		}
	}

	return false
}

func (a *Adapter) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (a *Adapter) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := a.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (a *Adapter) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{
		Success: true,
	}, nil
}
