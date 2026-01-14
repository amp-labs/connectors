package dixa

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	restAPIVersion = "v1"
	queues         = "queues"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	objectURL := url.String()

	if params.NextPage != "" {
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL)
		if err != nil {
			return nil, err
		}

		objectURL = url.String() + params.NextPage.String()
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, objectURL, nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		constructRecords(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

type request struct {
	Request any `json:"request"`
}

func constructQueuePayload(recordData any) request {
	return request{Request: recordData}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		payload = params.RecordData
		method  = http.MethodPost
	)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	if params.ObjectName == queues {
		payload = constructQueuePayload(params.RecordData)
	}

	jsonData, err := json.Marshal(payload)
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

	resp, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	recordId := parseRecordId(data)

	return &common.WriteResult{
		Success:  true,
		Data:     data,
		RecordId: recordId,
	}, nil
}

func parseRecordId(data map[string]any) string {
	v := getID(data)
	if v == nil {
		return ""
	}

	switch v := v.(type) {
	case string:
		return v
	case float64:
		return strconv.Itoa(int(v))
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

func getID(data map[string]any) any {
	for key, value := range data {
		if strings.EqualFold(key, "id") {
			return value
		}
	}

	return nil
}
