package teamleader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiListSuffix = ".list"
	pageSize      = 100
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	fullObjectName := objectName + apiListSuffix

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, fullObjectName)
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

	records, ok := (*res)["data"].([]any)
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
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	fullObjectName := params.ObjectName + apiListSuffix

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, fullObjectName)
	if err != nil {
		return nil, err
	}

	body := buildRequestBody(&params)

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(),
		nextRecordsURL(params),
		common.GetMarshaledData,
		params.Fields,
	)
}

func buildRequestBody(params *common.ReadParams) map[string]any {
	body := make(map[string]any)

	if !params.Since.IsZero() {
		body["filter"] = map[string]any{
			"updated_since": params.Since.Format(time.RFC3339),
		}
	}

	if params.NextPage != "" {
		body["page"] = map[string]any{
			"size":   pageSize,
			"number": params.NextPage,
		}
	} else {
		body["page"] = map[string]any{
			"size":   pageSize,
			"number": "1",
		}
	}

	return body
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	payload, ok := params.RecordData.(map[string]any)
	if !ok {
		return nil, errors.New("invalid record data") // nolint:err113
	}

	var fullObjectName string
	if params.RecordId != "" {
		fullObjectName = params.ObjectName + ".update"
		payload["id"] = params.RecordId
	} else {
		fullObjectName = writeFullObjectNames.Get(params.ObjectName)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, fullObjectName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
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

	dataNode, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(dataNode).StringRequired("id")
	if err != nil {
		return nil, err
	}

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

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}
