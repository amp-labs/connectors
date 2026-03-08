package restlet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	action := "create"
	if params.IsUpdate() {
		action = "update"
	}

	payload := writeRequest{
		Action: action,
		Type:   params.ObjectName,
	}

	if params.IsUpdate() {
		payload.RecordId = params.RecordId
	}

	// Extract values and sublists from RecordData.
	// If RecordData contains "values" key, use it; otherwise treat the entire map as values.
	if values, ok := recordData["values"]; ok {
		if valuesMap, ok := values.(map[string]any); ok {
			payload.Values = valuesMap
		}
	} else {
		payload.Values = recordData
	}

	if sublists, ok := recordData["sublists"]; ok {
		if sublistsMap, ok := sublists.(map[string]any); ok {
			payload.Sublists = sublistsMap
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal write request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.restletURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (a *Adapter) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	fullResp, err := common.UnmarshalJSON[restletResponse](resp)
	if err != nil {
		return nil, err
	}

	if fullResp.Header.Status != statusSuccess {
		return nil, parseRestletError(fullResp)
	}

	var writeBody writeResponseBody
	if err := json.Unmarshal(fullResp.Body, &writeBody); err != nil {
		return nil, fmt.Errorf("failed to parse write response body: %w", err)
	}

	recordId := fmt.Sprintf("%v", writeBody.RecordId)

	return &common.WriteResult{
		Success:  true,
		RecordId: recordId,
	}, nil
}
