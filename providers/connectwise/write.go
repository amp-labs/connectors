package connectwise

import (
	"context"
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

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		if updateIsPatchMode(params) {
			method = http.MethodPatch
		} else {
			method = http.MethodPut
		}
	}

	reader, err := params.GetRecordReader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), reader)
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

// updateIsPatchMode determines whether an update operation should use HTTP PATCH
// instead of PUT by checking if the payload is a JSON Patch array.
// This means payloads structured with "op", "path", "value"
// will trigger PATCH, while regular object payloads will use PUT.
//
// Example payload that triggers PATCH (JSON Patch):
//
//	[
//	  {"op": "replace", "path": "/firstName", "value": "Sims"},
//	]
//
// Example payload that uses PUT (regular object):
//
//	{
//	  "lastName": "Sims"
//	}
func updateIsPatchMode(params common.WriteParams) bool {
	payload, err := common.RecordDataToStruct[patchPayload](params)
	if err != nil {
		return false
	}

	if len(payload) == 0 {
		return false
	}

	if payload[0].Op != "" {
		return true
	}

	return false
}

type patchPayload []patchOperationPayload

type patchOperationPayload struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}
