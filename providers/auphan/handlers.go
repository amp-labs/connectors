package auphan

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	// Max page size for list endpoints.
	maxPageSize = 200
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	if objectCategoryReport.Has(objectName) {
		return c.newRequest(ctx, "reports/types.get", map[string]any{"type": objectName})
	}

	if list, ok := objectCategoryList[objectName]; ok {
		// One record is enough to sample the field names.
		return c.newRequest(ctx, list.endpoint, map[string]any{"page": 1, "per_page": 1})
	}

	return nil, common.ErrObjectNotSupported
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	// See parseReadResponse: a failure can arrive with HTTP 200.
	if err := validateResponseSuccess(response); err != nil {
		return nil, err
	}

	displayName := naming.CapitalizeFirstLetterEveryWord(naming.SeparateUnderscoreWords(objectName))

	switch {
	case objectCategoryReport.Has(objectName):
		return parseReportMetadata(displayName, response)
	case objectCategoryList.Has(objectName):
		return parseListMetadata(displayName, response)
	default:
		return nil, common.ErrObjectNotSupported
	}
}

// parseReportMetadata reads the columns reports/types.get lists for a report.
func parseReportMetadata(displayName string, response *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	reportType, err := common.UnmarshalJSON[reportTypeResponse](response)
	if err != nil {
		return nil, err
	}

	metadata := common.NewObjectMetadata(displayName, common.FieldsMetadata{})

	for _, column := range reportType.Data.Columns {
		metadata.AddFieldMetadata(column.ID, common.FieldMetadata{
			DisplayName:  column.Label,
			ValueType:    common.ValueTypeString, // values come back as strings, money included
			ProviderType: column.Kind,            // dimension or measure
		})
	}

	return metadata, nil
}

// parseListMetadata samples the field names from the first record of a list endpoint.
func parseListMetadata(displayName string, response *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	items, err := jsonquery.New(body, "data").ArrayOptional("items")
	if err != nil {
		return nil, err
	}

	records, err := jsonquery.Convertor.ArrayToMap(items)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrMissingMetadata
	}

	metadata := common.NewObjectMetadata(displayName, common.FieldsMetadata{})

	for field := range records[0] {
		metadata.AddFieldMetadata(field, common.FieldMetadata{
			DisplayName: field,
			ValueType:   common.ValueTypeOther,
		})
	}

	return metadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if objectCategoryReport.Has(params.ObjectName) {
		return c.buildReportRequest(ctx, params)
	}

	list, ok := objectCategoryList[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	page := 1

	if params.NextPage != "" {
		var err error

		page, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	perPage := min(readhelper.PageSizeWithDefault(params, maxPageSize), maxPageSize)

	return c.newRequest(ctx, list.endpoint, map[string]any{"page": page, "per_page": perPage})
}

func (c *Connector) buildReportRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	filters := c.reportFilters(params, objectCategoryReport[params.ObjectName])

	// Fields is a set, so its order changes from call to call. Sorting only keeps the request body
	// the same for the same fields; Auphan returns columns in the report's own order either way.
	columns := params.Fields.List()
	slices.Sort(columns)

	return c.newRequest(ctx, "reports/execute", map[string]any{
		"type":    params.ObjectName,
		"columns": columns,
		"filters": filters,
	})
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// Auphan can send a failure with HTTP 200: {"success": false, "error": {...}}.
	// The ErrorHandler only sees non-2xx responses, so such a failure reaches this point and is
	// caught here. Past this check the response is a success and only records are parsed.
	if err := validateResponseSuccess(response); err != nil {
		return nil, err
	}

	if objectCategoryReport.Has(params.ObjectName) {
		return common.ParseResult(
			response,
			extractReportRecords,
			reportNextPage,
			// Report records are totals grouped by dimensions, so they have no ID.
			readhelper.MakeMarshaledDataFuncWithId(mergeReportRecord, readhelper.IdFieldQuery{}),
			params.Fields,
		)
	}

	list, ok := objectCategoryList[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	return common.ParseResult(
		response,
		common.ExtractOptionalRecordsFromPath("items", "data"),
		nextPage,
		readhelper.MakeGetMarshaledDataWithId(list.idField),
		params.Fields,
	)
}

// newRequest builds a POST to {base}/{endpoint} with the {"store_id", "params"} body every call uses.
func (c *Connector) newRequest(ctx context.Context, endpoint string, params map[string]any) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, endpoint)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]any{
		"store_id": c.storeId,
		"params":   params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}
