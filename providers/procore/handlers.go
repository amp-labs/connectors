package procore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	defaultPageSize           = 1000
	headerProcoreCompanyID    = "Procore-Company-Id"
	queryParamPage            = "page"
	queryParamPerPage         = "per_page"
	queryParamUpdatedAtFilter = "filters[updated_at]"
	filterRangeSeparator      = "..."
)

var (
	ErrMissingCompanyID = errors.New("company metadata is required for this object")
	ErrInvalidObject    = errors.New("object name cannot be empty")
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.buildObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(queryParamPerPage, "1")

	return c.newRequest(ctx, http.MethodGet, url, nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateUnderscoreWords(objectName)),
	}

	records, err := extractRecords(response, objectName)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    analyzeValue(value),
			ProviderType: string(analyzeValue(value)),
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildObjectURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	page := "1"
	if params.NextPage != "" {
		page = params.NextPage.String()
	}

	url.WithQueryParam(queryParamPage, page)
	url.WithQueryParam(queryParamPerPage, strconv.Itoa(resolvePageSize(params.PageSize)))

	if objectRegistry[params.ObjectName].incremental {
		if filter := buildUpdatedAtFilter(params.Since, params.Until); filter != "" {
			url.WithQueryParam(queryParamUpdatedAtFilter, filter)
		}
	}

	return c.newRequest(ctx, http.MethodGet, url, nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// Procore paginates with a Link header, so we extract the next page token from there.
	linkHeader := response.Headers.Get("Link")

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(objectRegistry[params.ObjectName].recordsKey),
		nextPageFromLink(linkHeader),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.buildObjectURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	body, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal write payload: %w", err)
	}

	return c.newRequest(ctx, method, url, bytes.NewReader(body))
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

	// v2.0 endpoints wrap the record under a "data" key; v1.0 endpoints return it at the root.
	root := body
	recordsKey := objectRegistry[params.ObjectName].recordsKey

	if recordsKey != "" {
		obj, err := jsonquery.New(body).ObjectOptional(recordsKey)
		if err == nil && obj != nil {
			root = obj
		}
	}

	data, err := jsonquery.Convertor.ObjectToMap(root)
	if err != nil {
		return &common.WriteResult{Success: true}, nil //nolint:nilerr
	}

	recordID, err := jsonquery.New(root).TextWithDefault("id", "")
	if err != nil || recordID == "" {
		return &common.WriteResult{Success: true, Data: data}, nil //nolint:nilerr
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
