package fathom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	restAPIPrefix = "external"
	apiVersion    = "v1"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, apiVersion, objectName)
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

	records, ok := (*res)["items"].([]any)
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

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.ObjectName == "meetings" {
		// This will add query parameters to include additional data fields
		// in the meetings API response.
		addMeetingsQueryParams(url)
	}

	if !params.Since.IsZero() {
		url.WithQueryParam("created_after", params.Since.Format(time.RFC3339))
	}

	if !params.Until.IsZero() {
		url.WithQueryParam("created_before", params.Until.Format(time.RFC3339))
	}

	if params.NextPage != "" {
		url.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("items"),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	if params.RecordId != "" {
		method = http.MethodPut
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
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

	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     resp,
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
