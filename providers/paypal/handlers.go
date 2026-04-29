package paypal

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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
