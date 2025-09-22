package amplitude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiV2       = "2"
	apiV3       = "3"
	api2BaseURL = "https://api2.amplitude.com" // Base URL for Amplitude's HTTP API v2
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
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

	responseField := objectResponseField.Get(objectName)

	data, ok := (*res)[responseField].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := data[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	var (
		url *urlbuilder.URL
		err error
	)

	url, err = c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseKey := objectResponseField.Get(params.ObjectName)

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(responseKey),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) { //nolint:lll
	if api2SupportedObjects.Has(params.ObjectName) {
		return createRequestForApi2(ctx, c.HTTPClient().Client, params)
	}

	method := http.MethodPost

	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if supportedParamsPayloadObjectNames.Has(params.ObjectName) {
		return createRequestForParamsPayload(ctx, url, params)
	}

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func createRequestForApi2(ctx context.Context, client common.AuthenticatedHTTPClient, params common.WriteParams,
) (*http.Request, error) {
	apiKey, err := extractAPIKey(client, ctx)
	if err != nil {
		return nil, err
	}

	eventJSON, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	payloadKey := payloadKeyMapping.Get(params.ObjectName)

	form := make(url.Values)
	form.Add("api_key", apiKey)             // Use the extracted API key
	form.Add(payloadKey, string(eventJSON)) // Add as JSON string

	url, err := urlbuilder.New(api2BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func createRequestForParamsPayload(ctx context.Context, url *urlbuilder.URL, params common.WriteParams,
) (*http.Request, error) {
	recordMap, ok := params.RecordData.(map[string]interface{})
	if !ok {
		return nil, errors.New("params.RecordData is not a map[string]interface{}") //nolint: err113
	}

	for key, value := range recordMap {
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("annotation value for key %s is not a string", key) //nolint: err113
		}

		url.WithQueryParam(key, strValue)
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), nil)
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

	responseKey := writeObjectResponseField.Get(params.ObjectName)

	dataNode, err := jsonquery.New(body).ObjectOptional(responseKey)
	if err != nil {
		return nil, err
	}

	if dataNode == nil {
		// If object specific response key is not found, use the entire body
		dataNode = body
	}

	recordID, _ := jsonquery.New(dataNode).StrWithDefault("id", "")

	respMap, err := jsonquery.Convertor.ObjectToMap(dataNode)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     respMap,
	}, nil
}
