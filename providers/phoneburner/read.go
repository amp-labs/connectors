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
)

var (
	paginatedObjects = datautils.NewStringSet(
		"contacts",
		"dialsession",
		"members",
		"tags",
		"voicemails",
	)
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
		// Next page.
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

	// Many PhoneBurner list endpoints support page-based pagination. Defaulting these parameters
	// makes NextPage generation deterministic.
	if paginatedObjects.Has(params.ObjectName) {
		url.WithQueryParam("page_size", readhelper.PageSizeWithDefaultStr(params, "100"))
		url.WithQueryParam("page", "1")
	}

	// Apply time scoping when the provider supports it.
	switch params.ObjectName {
	case "contacts":
		// Docs: https://www.phoneburner.com/developer/route_list#contacts
		// updated_from / update_to in "YYYY-MM-DD HH:ii:ss" format.
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_from", params.Since.Format("2006-01-02 15:04:05"))
		}
		if !params.Until.IsZero() {
			url.WithQueryParam("update_to", params.Until.Format("2006-01-02 15:04:05"))
		}
	case "dialsession":
		// Docs: https://www.phoneburner.com/developer/route_list#dialsession
		// date_start / date_end in "YYYY-MM-DD" format.
		if !params.Since.IsZero() {
			url.WithQueryParam("date_start", params.Since.Format(time.DateOnly))
		}
		if !params.Until.IsZero() {
			url.WithQueryParam("date_end", params.Until.Format(time.DateOnly))
		}
	}

	return url, nil
}

func parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	_ = ctx

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
	case "members":
		if !params.Since.IsZero() || !params.Until.IsZero() {
			return common.ParseResultFiltered(
				params,
				response,
				common.MakeRecordsFunc("members", "members"),
				readhelper.MakeTimeFilterFunc(
					readhelper.ReverseOrder,
					readhelper.NewTimeBoundary(),
					"date_added",
					"2006-01-02 15:04:05",
					nextRecordsURL(url, params.ObjectName),
				),
				common.MakeMarshaledDataFunc(nil),
				params.Fields,
			)
		}
	case "voicemails":
		if !params.Since.IsZero() || !params.Until.IsZero() {
			return common.ParseResultFiltered(
				params,
				response,
				common.MakeRecordsFunc("voicemails", "voicemails"),
				readhelper.MakeTimeFilterFunc(
					readhelper.ReverseOrder,
					readhelper.NewTimeBoundary(),
					"created_when",
					"2006-01-02 15:04:05",
					nextRecordsURL(url, params.ObjectName),
				),
				common.MakeMarshaledDataFunc(nil),
				params.Fields,
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

	httpStatus := int(httpStatusI)
	if httpStatus == 0 {
		httpStatus = response.Code
	}

	if status != "" && status != "success" && httpStatus < 400 {
		httpStatus = http.StatusBadRequest
	}

	if !httpkit.Status2xx(httpStatus) || (status != "" && status != "success") {
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
			Header:     response.Headers,
		}, bodyBytes)
	}

	return nil
}

func recordsFunc(objectName string) (common.RecordsFunc, error) {
	switch objectName {
	// Docs: https://www.phoneburner.com/developer/route_list#contacts
	case "contacts":
		return common.ExtractRecordsFromPath("contacts", "contacts"), nil
	// Docs: https://www.phoneburner.com/developer/route_list#dialsession
	case "dialsession":
		return common.ExtractRecordsFromPath("dialsessions", "dialsessions"), nil
	// Docs: https://www.phoneburner.com/developer/route_list#members
	case "members":
		return common.ExtractRecordsFromPath("members", "members"), nil
	// Docs: https://www.phoneburner.com/developer/route_list#tags
	case "tags":
		return common.ExtractRecordsFromPath("tags", "tags"), nil
	// Docs: https://www.phoneburner.com/developer/route_list#voicemails
	case "voicemails":
		return common.ExtractRecordsFromPath("voicemails", "voicemails"), nil
	// Docs: https://www.phoneburner.com/developer/route_list#folders
	case "folders":
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
	case "dialsession":
		return "dialsessions"
	default:
		return objectName
	}
}

func extractFoldersRecords() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		foldersNode, err := jsonquery.New(node).ObjectRequired("folders")
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
