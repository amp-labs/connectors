package dropboxsign

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiVersion  = "v3"
	listSuffix  = "list"
	pageSizeKey = "page_size"
	pageSize    = "100"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, listSuffix)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	res, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	responseKey := readObjectResponseKey.Get(objectName)

	records, ok := (*res)[responseKey].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: string(inferValueTypeFromData(value)),
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, listSuffix)
	if err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		url.WithQueryParam("page", params.NextPage.String())
	} else {
		url.WithQueryParam("page", "1")
	}

	url.WithQueryParam(pageSizeKey, pageSize)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := readObjectResponseKey.Get(params.ObjectName)

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseKey),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := buildWriteURL(c.ProviderInfo().BaseURL, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

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
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	dataNode, err := jsonquery.New(body).ObjectOptional(params.ObjectName)
	if err != nil || dataNode == nil {
		return nil, err
	}

	responseKey := writeResponseKey.Get(params.ObjectName)

	recordId, err := jsonquery.New(dataNode).StrWithDefault(responseKey, "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(dataNode)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordId,
		Errors:   nil,
		Data:     resp,
	}, nil
}

func buildWriteURL(baseURL, objectName, recordId string) (string, error) {
	var urlSuffix string

	// For objects that do not require 'create' suffix on write without record ID
	// e.g., ApiApp, we skip adding the 'create' suffix.
	if recordId == "" && !writeObjectWithoutCreateSuffix.Has(objectName) {
		urlSuffix = "create"
	}

	url, err := urlbuilder.New(baseURL, apiVersion, objectName, urlSuffix)
	if err != nil {
		return "", err
	}

	// For objects that require update by ID on write with record ID
	// e.g., ApiApp, we append the record ID to the URL.
	if recordId != "" && writeObjectUpdateById.Has(objectName) {
		url.AddPath(recordId)
	}

	return url.String(), nil
}
