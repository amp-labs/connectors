// Package connectwise implements the ConnectWise provider connector, handling
// object reads, writes, and provider-specific quirks (custom fields,
// communication items, JSON Patch behavior, etc.).
package connectwise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
	"github.com/amp-labs/connectors/providers/connectwise/internal/customfields"
)

// buildWriteRequest constructs an HTTP request for a write operation (create or update) against a ConnectWise object.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.GetUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	data, method, err := c.makeWritePayload(ctx, params, url)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	c.ClientIdHeader().ApplyToRequest(req)

	return req, nil
}

// makeWritePayload builds the write payload and HTTP method for a create or
// update operation, and adjusts the URL as needed.
func (c *Connector) makeWritePayload(ctx context.Context,
	params common.WriteParams,
	url *urlbuilder.URL,
) (any, string, error) {
	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		return c.makeWriteUpdatePayload(ctx, params)
	}

	return c.makeWriteCreatePayload(ctx, params)
}

// makeWriteCreatePayload builds a create (POST) payload for the given object.
//
// It:
//   - Extracts the record from params.
//   - Applies custom-field normalization via payloadWithCustomFields.
//   - For contacts, enriches the record with communication items by translating
//     virtual fields into a proper `communicationItems` array.
//
// Returns the record, http.MethodPost, and any error encountered.
func (c *Connector) makeWriteCreatePayload(ctx context.Context, params common.WriteParams) (any, string, error) {
	record, err := params.GetRecord()
	if err != nil {
		return nil, "", err
	}

	customfields.TransformCreatePayload(record)

	if params.ObjectName == objectNameContacts {
		if err = c.postPayloadWithCommunicationItems(ctx, record); err != nil {
			return nil, "", err
		}
	}

	return record, http.MethodPost, nil
}

// makeWriteUpdatePayload builds the update payload and HTTP method for the
// given object.
//
// params.RecordData holds the update payload. Depending on its shape:
//   - JSON Patch array -> HTTP PATCH (with custom-field normalization).
//   - Regular object -> HTTP PUT, except for contacts.
//
// ConnectWise specifics:
//   - PATCH body must be a bare JSON array of operations.
//   - For contacts, full replacement is implemented via PATCH instead of PUT
//     to avoid known bugs with `communicationItems` validation.
func (c *Connector) makeWriteUpdatePayload(ctx context.Context, params common.WriteParams) (any, string, error) {
	// Detect JSON Patch payloads and use HTTP PATCH.
	if payload, ok := extractPatchPayload(params); ok {
		data := customfields.TransformUpdatePayload(params.ObjectName, payload.Patch)

		if params.ObjectName == objectNameContacts {
			items, err := c.contactsPartialUpdatePayload(ctx, data, params.RecordId)

			return items, http.MethodPatch, err
		}

		return data, http.MethodPatch, nil
	}

	// Regular object payload.
	record, err := params.GetRecord()
	if err != nil {
		return nil, "", err
	}

	customfields.TransformCreatePayload(record)

	if params.ObjectName == objectNameContacts {
		// For contacts, emulate full PUT via PATCH to avoid ConnectWise bugs.
		items, err := c.contactsFullUpdatePayload(ctx, record, params.RecordId)

		return items, http.MethodPatch, err
	}

	return record, http.MethodPut, nil
}

// parseWriteResponse parses the HTTP response from a write operation and
// constructs a common.WriteResult.
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

// extractPatchPayload detects whether an update operation should use HTTP PATCH
// instead of PUT by inspecting the payload structure.
//
// It interprets the incoming RecordData as a patchPayload:
//   - If the `patch` field is non-empty and its first element has a non-empty
//     `Op` (indicating a JSON Patch operation like "add", "replace", "remove"),
//     the function returns the parsed payload and true.
//   - Otherwise, it returns nil, false, indicating a regular object payload
//     suitable for PUT.
//
// Expected JSON Patch-style payload (triggers PATCH):
//
//	{
//	  "patch": [
//	    {"op": "replace", "path": "/firstName", "value": "Sims"}
//	  ]
//	}
//
// Expected regular object payload (triggers PUT):
//
//	{
//	  "lastName": "Sims"
//	}
//
// This function encapsulates the connector's convention for distinguishing
// between partial (JSON Patch) and full (PUT) updates.
func extractPatchPayload(params common.WriteParams) (*core.PatchPayload, bool) {
	payload, err := common.RecordDataToStruct[core.PatchPayload](params)
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
