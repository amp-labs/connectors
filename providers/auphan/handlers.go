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
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	// Max page size for list endpoints.
	pageSize = 200
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	if reportObjects.Has(objectName) {
		return c.newRequest(ctx, "reports/types.get", map[string]any{"type": objectName})
	}

	endpoint, ok := listObjects[objectName]
	if !ok {
		return nil, common.ErrObjectNotSupported
	}

	// One record is enough to sample the field names.
	return c.newRequest(ctx, endpoint, map[string]any{"page": 1, "per_page": 1})
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	if err := validateResponseSuccess(response); err != nil {
		return nil, err
	}

	displayName := naming.CapitalizeFirstLetterEveryWord(naming.SeparateUnderscoreWords(objectName))

	metadata := common.NewObjectMetadata(displayName, common.FieldsMetadata{})

	if reportObjects.Has(objectName) {
		reportType, err := common.UnmarshalJSON[reportTypeResponse](response)
		if err != nil {
			return nil, err
		}

		for _, column := range reportType.Data.Columns {
			metadata.AddFieldMetadata(column.ID, common.FieldMetadata{
				DisplayName:  column.Label,
				ValueType:    common.ValueTypeString, // values come back as strings, money included
				ProviderType: column.Kind,            // dimension or measure
			})
		}

		return metadata, nil
	}

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

	for field := range records[0] {
		metadata.AddFieldMetadata(field, common.FieldMetadata{
			DisplayName: field,
			ValueType:   common.ValueTypeOther,
		})
	}

	return metadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if reportObjects.Has(params.ObjectName) {
		return c.buildReportRequest(ctx, params)
	}

	endpoint, ok := listObjects[params.ObjectName]
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

	return c.newRequest(ctx, endpoint, map[string]any{"page": page, "per_page": pageSize})
}

func (c *Connector) buildReportRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	filters := c.reportFilters(params, reportObjects[params.ObjectName])

	// Fields is a set, so sort for a stable request. Auphan ids are lowercase.
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
	if err := validateResponseSuccess(response); err != nil {
		return nil, err
	}

	if reportObjects.Has(params.ObjectName) {
		return parseReport(response, params.Fields.List())
	}

	return common.ParseResult(
		response,
		common.ExtractOptionalRecordsFromPath("items", "data"),
		nextPage,
		common.GetMarshaledData,
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
