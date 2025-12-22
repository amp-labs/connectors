package cloudtalk

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/cloudtalk/metadata"
	"github.com/spyzhov/ajson"
)

type Response struct {
	ResponseData ResponseData `json:"responseData"`
}

type ResponseData struct {
	Data       []map[string]any `json:"data"`
	PageNumber int              `json:"pageNumber"`
	PageCount  int              `json:"pageCount"`
	Limit      int              `json:"limit"`
	ItemsCount int              `json:"itemsCount"`
}

type ResponseError struct {
	Error string `json:"error"`
}

func (r *ResponseError) CombineErr(baseErr error) error {
	if r.Error == "" {
		return baseErr
	}

	return fmt.Errorf("%w: %s", baseErr, r.Error)
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := c.getURLPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	// Pagination
	// CloudTalk uses 1-based indexing for pages
	page := "1"
	if params.NextPage != "" {
		page = params.NextPage.String()
	}

	url.WithQueryParam("page", page)

	if params.PageSize > 0 {
		url.WithQueryParam("limit", strconv.Itoa(params.PageSize))
	}

	// Filtering
	if params.ObjectName == "calls" || params.ObjectName == "activity" {
		// https://developers.cloudtalk.io/reference/list-calls#tag/Calls
		// https://developers.cloudtalk.io/reference/list-activities
		if !params.Since.IsZero() {
			url.WithQueryParam("date_from", params.Since.Format(time.DateTime))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("date_to", params.Until.Format(time.DateTime))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) getURLPath(objectName string) (string, error) {
	return metadata.Schemas.FindURLPath(c.ProviderContext.Module(), objectName)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("data", "responseData"),
		makeFilterFunc(params),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

func makeFilterFunc(_ common.ReadParams) common.RecordsFilterFunc {
	// CloudTalk only supports provider-side filtering for 'calls' and 'activity'.
	// For other objects, the provider does not expose a timestamp field in the list response,
	// so client-side filtering is not possible.
	// We return an identity filter to perform a full read.
	return readhelper.MakeIdentityFilterFunc(makeNextPageToken)
}

func makeNextPageToken(node *ajson.Node) (string, error) {
	responseData := jsonquery.New(node, "responseData")

	pageNumber, err := responseData.IntegerWithDefault("pageNumber", 0)
	if err != nil {
		return "", err
	}

	pageCount, err := responseData.IntegerWithDefault("pageCount", 0)
	if err != nil {
		return "", err
	}

	var nextPage string
	if pageNumber < pageCount {
		// CloudTalk uses 1-based indexing for pages usually, and returns current pageNumber.
		// Next page is simply current + 1
		nextPage = strconv.Itoa(int(pageNumber) + 1)
	}

	return nextPage, nil
}

var (
	errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
		interpreter.FormatTemplate{
			MustKeys: []string{"error"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	)

	statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
		http.StatusNotFound: common.ErrNotFound,
	}

	interpretError = interpreter.ErrorHandler{ // nolint:gochecknoglobals
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle
)
