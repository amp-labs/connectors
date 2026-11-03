package restlet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	// Always inject type from the URL's object name if not already in the record.
	if _, hasType := recordData["type"]; !hasType {
		recordData["type"] = params.ObjectName
	}

	// [ENG-3740] If the server already set an action (e.g. "transform", "void"), use it as-is.
	// This supports NetSuite-specific operations like record transformation and voiding.
	// Otherwise infer create/update and inject action/recordId.
	//
	// Pass the record data through to the RESTlet as-is, just inject action/recordId.
	// This allows callers to send any RESTlet-supported keys (values, textValues, sublists,
	// subrecords, defaultValues, options, etc.) without the connector needing to know about each one.
	if _, hasAction := recordData["action"]; !hasAction {
		action := "create"
		if params.IsUpdate() {
			action = "update"
		}

		recordData["action"] = action

		if params.IsUpdate() {
			recordData["recordId"] = params.RecordId
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

	node, err := ajson.Unmarshal(fullResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse write response body: %w", err)
	}

	rawID, err := jsonquery.New(node).TextWithDefault("recordId", "")
	if err != nil {
		//nolint:nilerr
		return withRestletVersion(&common.WriteResult{Success: true}, fullResp.Header), nil
	}

	result := &common.WriteResult{
		Success:  true,
		RecordId: rawID,
	}

	if data, err := jsonquery.Convertor.ObjectToMap(node); err == nil {
		result.Data = data
	}

	return withRestletVersion(result, fullResp.Header), nil
}

// restletVersionKey is where the RESTlet's version goes in WriteResult.Data, so callers can tell
// which bundle release each NetSuite account runs. The RESTlet sends it in the response header,
// which the write result doesn't otherwise carry.
const restletVersionKey = "_scriptVersion"

func withRestletVersion(result *common.WriteResult, header responseHeader) *common.WriteResult {
	version := header.restletVersion()
	if version == "" {
		return result
	}

	if result.Data == nil {
		result.Data = map[string]any{}
	}

	result.Data[restletVersionKey] = version

	return result
}
