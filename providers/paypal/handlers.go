package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

type timeFilter struct {
	sinceParam string
	untilParam string
}

// objectMaxPageSize maps each object that accepts a page_size query param to its API maximum.
//
//nolint:gochecknoglobals,mnd
var objectMaxPageSize = map[string]int{
	"disputes":        50,
	"invoices":        100,
	"templates":       100,
	"plans":           20,
	"products":        20,
	"transactions":    500,
	"webhooks-events": 50,
}

//nolint:gochecknoglobals
var objectTimeFilter = datautils.NewDefaultMap(
	map[string]timeFilter{
		"disputes":        {sinceParam: "update_time_after", untilParam: "update_time_before"},
		"transactions":    {sinceParam: "start_date", untilParam: "end_date"},
		"webhooks-events": {sinceParam: "start_time", untilParam: "end_time"},
		"balances":        {sinceParam: "as_of_time", untilParam: ""},
	},
	func(_ string) timeFilter {
		return timeFilter{}
	},
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) { //nolint:cyclop,lll
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	path, err := schemas.FindURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if maxSize, ok := objectMaxPageSize[params.ObjectName]; ok {
		pageSize := maxSize
		if params.PageSize > 0 && params.PageSize < maxSize {
			pageSize = params.PageSize
		}

		url.WithQueryParam("page_size", strconv.Itoa(pageSize))
	}

	tf := objectTimeFilter.Get(params.ObjectName)

	if tf.sinceParam != "" && !params.Since.IsZero() {
		url.WithQueryParam(tf.sinceParam, params.Since.Format(time.RFC3339))
	}

	if tf.untilParam != "" && !params.Until.IsZero() {
		url.WithQueryParam(tf.untilParam, params.Until.Format(time.RFC3339))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := schemas.LookupArrayFieldName(common.ModuleRoot, params.ObjectName)

	return common.ParseResult(
		response,
		common.ExtractOptionalRecordsFromPath(responseKey),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

// objectUpdateMethod maps objects to their HTTP update method. Defaults to PATCH.
//
//nolint:gochecknoglobals
var objectUpdateMethod = datautils.NewDefaultMap(
	map[string]string{
		"invoices":     http.MethodPut,
		"templates":    http.MethodPut,
		"web-profiles": http.MethodPut,
	},
	func(_ string) string {
		return http.MethodPatch
	},
)

// objectWritePath holds write paths for objects that are not in the read schema (write-only).
//
//nolint:gochecknoglobals
var objectWritePath = map[string]string{
	"orders": "/v2/checkout/orders",
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	path, err := schemas.FindURLPath(common.ModuleRoot, params.ObjectName)
	if err != nil {
		var ok bool

		path, ok = objectWritePath[params.ObjectName]
		if !ok {
			return nil, err
		}
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)
		method = objectUpdateMethod.Get(params.ObjectName)
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	recordID, err := jsonquery.New(body).StringOptional("id")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	id := ""
	if recordID != nil {
		id = *recordID
	}

	return &common.WriteResult{
		RecordId: id,
		Success:  true,
		Data:     resp,
	}, nil
}
