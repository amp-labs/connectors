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

	// [ENG-3740] If the server already set an action (e.g. "transform"), use it as-is.
	// This supports NetSuite-specific operations like record transformation.
	// Otherwise infer create/update and inject action/type/recordId.
	//
	// Pass the record data through to the RESTlet as-is, just inject action/type/recordId.
	// This allows callers to send any RESTlet-supported keys (values, textValues, sublists,
	// subrecords, defaultValues, options, etc.) without the connector needing to know about each one.
	if _, hasAction := recordData["action"]; !hasAction {
		action := "create"
		if params.IsUpdate() {
			action = "update"
		}

		recordData["action"] = action
		recordData["type"] = params.ObjectName

		if params.IsUpdate() {
			recordData["recordId"] = params.RecordId
		}
	} else {
		// Action already set (e.g., transform, void) — inject type from URL if not in record
		if _, hasType := recordData["type"]; !hasType {
			recordData["type"] = params.ObjectName
		}
	}

	body, err := json.Marshal(recordData)
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
