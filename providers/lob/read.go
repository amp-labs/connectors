package lob

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Endpoints state 100 as an upper limit.
// Exceeding such will result in error with
// status code 422 and message: `limit must be less than or equal to 100`.
// Example: https://docs.lob.com/#tag/Addresses/operation/addresses_list
const defaultPageSize = "100"

// Objects that use time based query param for record filtering.
// Note: many objects could be filtered using `date_created`.
var incrementalReadByProvider = map[string]string{ // nolint:gochecknoglobals
	"billing_groups": "date_modified",
}

// Objects that use connector side filtering of records.
//
// Every object has `date_modified` in response body.
//
// Note: Some objects could be filtered using provider API via `date_created`
// but this field is much more restrictive compared to the `date_modified`.
// Therefore, in order to surface up-to-date changes connector side filtering is favored.
var incrementalReadByConnector = datautils.NewSetFromList([]string{ // nolint:gochecknoglobals
	"addresses",
	"bank_accounts",
	"booklets",
	"buckslips",
	"campaigns",
	"cards",
	"checks",
	"informed_delivery_campaigns",
	"letters",
	"postcards",
	"self_mailers",
	"snap_packs",
	"templates",
})

var readResponseArrayField = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
	"update": "", // root level is an array.
}, func(objectName string) string {
	return "data" // most objects have records under `data` key.
})

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.GetURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	if err = attachReadFilterQueryParams(url, params); err != nil {
		return nil, err
	}

	return url, nil
}

func attachReadFilterQueryParams(url *urlbuilder.URL, params common.ReadParams) error {
	queryParam, ok := incrementalReadByProvider[params.ObjectName]
	if !ok {
		// No-op, no time based query params exist
		return nil
	}

	// https://docs.lob.com/#tag/Billing-Groups/operation/billing_groups_list
	// Query value is a JSON type string.
	queryValue := map[string]string{}

	if !params.Since.IsZero() {
		queryValue["gt"] = datautils.Time.FormatRFC3339inUTC(params.Since)
	}

	if !params.Until.IsZero() {
		queryValue["lt"] = datautils.Time.FormatRFC3339inUTC(params.Until)
	}

	if len(queryValue) == 0 {
		// Since and Until are not specified in params.
		return nil
	}

	// Need to pass the query param.
	value, err := json.Marshal(queryValue)
	if err != nil {
		return err
	}

	url.WithQueryParam(queryParam, string(value))

	return nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := readResponseArrayField.Get(params.ObjectName)

	return common.ParseResultFiltered(
		params,
		response,
		makeGetRecords(responseFieldName),
		makeFilterFunc(params),
		readhelper.MakeMarshaledSelectedDataFunc(
			fieldsSelector,
			jsonquery.Convertor.ObjectToMap,
		),
		params.Fields,
	)
}

func makeGetRecords(responseFieldName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

func makeFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	nextPageFunc := makeNextRecordsURL()

	if incrementalReadByConnector.Has(params.ObjectName) {
		return readhelper.MakeTimeFilterFunc(
			readhelper.ReverseOrder,
			readhelper.NewTimeBoundary(),
			"date_modified",
			time.RFC3339,
			nextPageFunc,
		)
	}

	// Default no filtering.
	return readhelper.MakeIdentityFilterFunc(nextPageFunc)
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node).StrWithDefault("next_url", "")
	}
}

func fieldsSelector(node *ajson.Node, fields []string) (map[string]any, string, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, "", err
	}

	identifier, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, "", err
	}

	selected := readhelper.SelectFields(root, datautils.NewSetFromList(fields))

	return selected, identifier, nil
}
