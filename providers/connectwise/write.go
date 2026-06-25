package connectwise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	// By contract, WriteParams.RecordData holds the JSON payload.
	// For updates, we may need to switch to PATCH (JSON Patch) or use PUT (replace).
	data := params.RecordData
	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		// If the incoming WriteParams.RecordData represents a JSON Patch payload, we must use HTTP PATCH.
		// Important: the connector's internal contract requires RecordData
		// to be a JSON object, not a top-level JSON array.
		// However, ConnectWise provider expects a PATCH body as a bare JSON array.
		if payload, ok := extractPatchPayload(params); ok {
			method = http.MethodPatch
			data = payload.Patch
		} else {
			method = http.MethodPut
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	c.clientIdHeader().ApplyToRequest(req)

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

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

// extractPatchPayload determines whether an update operation should use HTTP PATCH
// instead of PUT by checking if the payload is a JSON Patch array.
// This means payloads structured with "op", "path", "value"
// will trigger PATCH, while regular object payloads will use PUT.
//
// Example payload that triggers PATCH (JSON Patch):
//
//	"patch": [
//	  {"op": "replace", "path": "/firstName", "value": "Sims"},
//	]
//
// Example payload that uses PUT (regular object):
//
//	{
//	  "lastName": "Sims"
//	}
func extractPatchPayload(params common.WriteParams) (*patchPayload, bool) {
	payload, err := common.RecordDataToStruct[patchPayload](params)
	if err != nil {
		return nil, false
	}

	if len(payload.Patch) == 0 {
		return nil, false
	}

	if payload.Patch[0].Op != "" {
		return &payload, true
	}

	return nil, false
}

type patchPayload struct {
	Patch []patchOperationPayload `json:"patch"`
}

type patchOperationPayload struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}
