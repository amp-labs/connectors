package phoneburner

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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
		"customfields",
		"dialsession",
		"members",
		"voicemails",
	)
)

func buildReadRequest(ctx context.Context, baseURL string, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		// Next page.
		url, err := urlbuilder.New(params.NextPage.String())
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

	url, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Many PhoneBurner list endpoints support page-based pagination. Defaulting these parameters
	// makes NextPage generation deterministic.
	if paginatedObjects.Has(params.ObjectName) {
		url.WithQueryParam("page_size", strconv.Itoa(100))
		url.WithQueryParam("page", "1")
	}

	// Apply time scoping when the provider supports it.
	switch params.ObjectName {
	case "contacts":
		// Docs: updated_from / update_to in "YYYY-MM-DD HH:ii:ss" format.
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_from", params.Since.Format("2006-01-02 15:04:05"))
		}
		if !params.Until.IsZero() {
			url.WithQueryParam("update_to", params.Until.Format("2006-01-02 15:04:05"))
		}
	case "dialsession":
		// Docs: date_start / date_end in "YYYY-MM-DD" format.
		if !params.Since.IsZero() {
			url.WithQueryParam("date_start", params.Since.Format(time.DateOnly))
		}
		if !params.Until.IsZero() {
			url.WithQueryParam("date_end", params.Until.Format(time.DateOnly))
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
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
	case "contacts":
		// Contacts can include a "custom_fields" array.
		return common.ParseResult(
			response,
			common.MakeRecordsFunc("contacts", "contacts"),
			nextRecordsURL(url.String(), params.ObjectName),
			common.MakeMarshaledDataFunc(flattenContactCustomFields),
			params.Fields,
		)
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
					nextRecordsURL(url.String(), params.ObjectName),
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
					nextRecordsURL(url.String(), params.ObjectName),
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
		nextRecordsURL(url.String(), params.ObjectName),
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

	if httpStatus >= 400 || (status != "" && status != "success") {
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
	case "contacts":
		return common.ExtractRecordsFromPath("contacts", "contacts"), nil
	case "customfields":
		return common.ExtractRecordsFromPath("customfields", "customfields"), nil
	case "dialsession":
		return common.ExtractRecordsFromPath("dialsessions", "dialsessions"), nil
	case "members":
		return common.ExtractRecordsFromPath("members", "members"), nil
	case "voicemails":
		return common.ExtractRecordsFromPath("voicemails", "voicemails"), nil
	case "folders":
		return extractFoldersRecords(), nil
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

func nextRecordsURL(requestURL string, objectName string) common.NextPageFunc {
	if !paginatedObjects.Has(objectName) {
		return func(*ajson.Node) (string, error) { return "", nil }
	}

	return func(node *ajson.Node) (string, error) {
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

		nextURL, err := urlbuilder.New(requestURL)
		if err != nil {
			return "", err
		}

		nextURL.WithQueryParam("page", strconv.Itoa(int(page)+1))

		return nextURL.String(), nil
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

		// Map iteration order is non-deterministic; keep output stable for tests and consumers.
		sort.Slice(out, func(i, j int) bool {
			ai, _ := out[i]["folder_id"].(string)
			aj, _ := out[j]["folder_id"].(string)
			return ai < aj
		})

		return out, nil
	}
}
