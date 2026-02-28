package phoneburner

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// API reference:
	// https://www.phoneburner.com/developer/route_list
	restPrefix = "rest"
	restVer    = "1"

	objectContacts    = "contacts"
	objectMembers     = "members"
	objectFolders     = "folders"
	objectDialsession = "dialsession"
	objectTags        = "tags"
	objectVoicemails  = "voicemails"
)

var paginatedObjects = datautils.NewStringSet( //nolint:gochecknoglobals
	objectContacts,
	objectDialsession,
	objectMembers,
	objectTags,
	objectVoicemails,
)

func buildReadRequest(ctx context.Context, baseURL string, params common.ReadParams) (*http.Request, error) {
	url, err := buildReadURL(baseURL, params)
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

func buildReadURL(baseURL string, params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	if params.ObjectName == "" {
		return nil, common.ErrMissingObjects
	}

	// Validate object support early (avoid issuing requests for unsupported objects).
	if _, err := recordsFunc(params.ObjectName); err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if paginatedObjects.Has(params.ObjectName) {
		url.WithQueryParam("page_size", readhelper.PageSizeWithDefaultStr(params, "100"))
		url.WithQueryParam("page", "1")
	}

	applyTimeScopingToURL(url, params)

	return url, nil
}

// phoneBurnerPST is PST (UTC-8). PhoneBurner compares timestamps in this timezone,
// so all time params must be converted to PST before formatting.
var phoneBurnerPST = time.FixedZone("PST", -8*60*60) //nolint:gochecknoglobals

// applyTimeScopingToURL adds object-specific time-filter query params.
func applyTimeScopingToURL(url *urlbuilder.URL, params common.ReadParams) {
	switch params.ObjectName {
	case objectContacts:
		// Docs: https://www.phoneburner.com/developer/route_list#contacts
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_from", params.Since.In(phoneBurnerPST).Format("2006-01-02 15:04:05"))
			url.WithQueryParam("include_new", "1")

			// Always send update_to explicitly; omitting it lets PhoneBurner default
			// to PST "now", which can be earlier than updated_from.
			updateTo := time.Now().In(phoneBurnerPST).Add(24 * time.Hour)
			if !params.Until.IsZero() {
				updateTo = params.Until.In(phoneBurnerPST)
			}

			url.WithQueryParam("update_to", updateTo.Format("2006-01-02 15:04:05"))
		}
	case objectDialsession:
		// Docs: https://www.phoneburner.com/developer/route_list#dialsession
		// date_start / date_end in "YYYY-MM-DD" format.
		if !params.Since.IsZero() {
			url.WithQueryParam("date_start", params.Since.Format(time.DateOnly))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("date_end", params.Until.Format(time.DateOnly))
		}
	}
}

func parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// PhoneBurner sometimes encodes errors in a 2xx response body using an envelope like:
	// { "http_status": 401, "status": "error", ... }
	// Convert these "200-with-error" responses into proper HTTP errors.
	if err := interpretPhoneBurnerEnvelopeError(response); err != nil {
		return nil, err
	}

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	switch params.ObjectName {
	case objectMembers:
		if !params.Since.IsZero() || !params.Until.IsZero() {
			return parseFilteredObjectResponse(params, response, objectMembers, "date_added", nextRecordsURL(url, objectMembers))
		}
	case objectVoicemails:
		if !params.Since.IsZero() || !params.Until.IsZero() {
			return parseFilteredObjectResponse(
				params, response, objectVoicemails, "created_when", nextRecordsURL(url, objectVoicemails),
			)
		}
	}

	records, err := recordsFunc(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		records,
		nextRecordsURL(url, params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

// parseFilteredObjectResponse handles time-filtered reads for objects that do not natively
// support server-side date filtering (members, voicemails).
func parseFilteredObjectResponse(
	params common.ReadParams,
	response *common.JSONHTTPResponse,
	objectName, timeField string,
	nextPage common.NextPageFunc,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(
		params,
		response,
		common.MakeRecordsFunc(objectName, objectName),
		readhelper.MakeTimeFilterFunc(
			readhelper.ReverseOrder,
			readhelper.NewTimeBoundary(),
			timeField,
			"2006-01-02 15:04:05",
			nextPage,
		),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

func interpretPhoneBurnerEnvelopeError(response *common.JSONHTTPResponse) error {
	body, ok := response.Body()
	if !ok {
		return nil
	}

	q := jsonquery.New(body)

	status, err := q.StrWithDefault("status", "")
	if err != nil {
		return err
	}

	httpStatusI, err := q.IntegerWithDefault("http_status", int64(response.Code))
	if err != nil {
		return err
	}

	httpStatus := resolveHTTPStatus(response.Code, status, int(httpStatusI))

	if !httpkit.Status2xx(httpStatus) || (status != "" && status != "success") {
		return buildEnvelopeHTTPError(body, httpStatus, response.Headers)
	}

	return nil
}

// resolveHTTPStatus normalises the PhoneBurner envelope status code.
func resolveHTTPStatus(responseCode int, status string, rawStatus int) int {
	httpStatus := rawStatus
	if httpStatus == 0 {
		httpStatus = responseCode
	}

	if status != "" && status != "success" && httpStatus < 400 {
		httpStatus = http.StatusBadRequest
	}

	return httpStatus
}

// buildEnvelopeHTTPError marshals the envelope body into a standard connector error.
func buildEnvelopeHTTPError(body *ajson.Node, httpStatus int, headers http.Header) error {
	raw, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return common.InterpretError(&http.Response{
		StatusCode: httpStatus,
		Header:     headers,
	}, bodyBytes)
}

func recordsFunc(objectName string) (common.RecordsFunc, error) {
	switch objectName {
	// Docs: https://www.phoneburner.com/developer/route_list#contacts
	case objectContacts:
		return common.ExtractRecordsFromPath(objectContacts, objectContacts), nil
	// Docs: https://www.phoneburner.com/developer/route_list#dialsession
	case objectDialsession:
		return common.ExtractRecordsFromPath("dialsessions", "dialsessions"), nil
	// Docs: https://www.phoneburner.com/developer/route_list#members
	case objectMembers:
		return common.ExtractRecordsFromPath(objectMembers, objectMembers), nil
	// Docs: https://www.phoneburner.com/developer/route_list#tags
	case objectTags:
		return common.ExtractRecordsFromPath(objectTags, objectTags), nil
	// Docs: https://www.phoneburner.com/developer/route_list#voicemails
	case objectVoicemails:
		return common.ExtractRecordsFromPath(objectVoicemails, objectVoicemails), nil
	// Docs: https://www.phoneburner.com/developer/route_list#folders
	case objectFolders:
		// Note: folders payload is an object-of-objects (not a JSON array), so we must flatten it.
		return extractFoldersRecords(), nil
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

func nextRecordsURL(requestURL *urlbuilder.URL, objectName string) common.NextPageFunc {
	if !paginatedObjects.Has(objectName) {
		return func(*ajson.Node) (string, error) { return "", nil }
	}

	return func(node *ajson.Node) (string, error) {
		if requestURL == nil {
			return "", nil
		}

		wrapper, err := jsonquery.New(node).ObjectRequired(paginationWrapperKey(objectName))
		if err != nil {
			return "", err
		}

		page, err := jsonquery.New(wrapper).IntegerWithDefault("page", 0)
		if err != nil {
			return "", err
		}

		totalPages, err := jsonquery.New(wrapper).IntegerWithDefault("total_pages", 0)
		if err != nil {
			return "", err
		}

		if totalPages == 0 || page >= totalPages {
			return "", nil
		}

		requestURL.WithQueryParam("page", strconv.Itoa(int(page)+1))

		return requestURL.String(), nil
	}
}

func paginationWrapperKey(objectName string) string {
	switch objectName {
	case objectDialsession:
		return "dialsessions"
	default:
		return objectName
	}
}

func extractFoldersRecords() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		foldersNode, err := jsonquery.New(node).ObjectRequired(objectFolders)
		if err != nil {
			return nil, err
		}

		m, err := jsonquery.Convertor.ObjectToMap(foldersNode)
		if err != nil {
			return nil, err
		}

		out := make([]map[string]any, 0, len(m))

		for _, v := range m {
			obj, ok := v.(map[string]any)
			if !ok || obj == nil {
				continue
			}

			out = append(out, obj)
		}

		return out, nil
	}
}
