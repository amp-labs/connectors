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
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// buildWriteRequest constructs an HTTP request for a write operation (create or update) against a ConnectWise object.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getURL(params.ObjectName)
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

	c.clientIdHeader().ApplyToRequest(req)

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

	payloadWithCustomFields(record)

	if params.ObjectName == objectNameContacts {
		if err = c.payloadWithCommunicationItems(ctx, record); err != nil {
			return nil, "", err
		}
	}

	return record, http.MethodPost, nil
}

// makeWriteUpdatePayload builds an update payload for the given object.
//
// By contract, params.RecordData holds the JSON payload for the update.
// For updates, we may:
//   - Use HTTP PATCH with a JSON Patch array when the caller provides a
//     patch-style payload.
//   - Use HTTP PUT (full replacement) for regular object payloads.
//
// Behavior:
//   - If the payload is detected as a JSON Patch (via extractPatchPayload),
//     the function:
//   - Normalizes custom fields for the patch (payloadPatchWithCustomFields).
//   - For contacts, translates virtual communication-item fields into
//     concrete JSON Patch operations targeting `communicationItems`.
//   - Returns the resulting patch array and http.MethodPatch.
//   - Otherwise, it:
//   - Extracts the record, applies custom-field normalization.
//   - For contacts with `communicationItems`, works around a ConnectWise
//     validation bug by clearing `communicationItems` via a preliminary
//     PATCH before performing the PUT.
//   - Returns the record and http.MethodPut.
//
// ConnectWise-specific notes:
//   - PATCH body must be a bare JSON array of operations, not wrapped in an
//     object.
//   - Contacts: including `communicationItems` in a PUT can trigger spurious
//     validation errors; we clear them out first via PATCH.
func (c *Connector) makeWriteUpdatePayload(ctx context.Context, params common.WriteParams) (any, string, error) {
	// Detect JSON Patch payloads and use HTTP PATCH if present.
	if payload, ok := extractPatchPayload(params); ok {
		data := payloadPatchWithCustomFields(params.ObjectName, payload.Patch)

		if params.ObjectName == objectNameContacts {
			items, err := c.patchPayloadWithCommunicationItems(ctx, data, params.RecordId)

			return items, http.MethodPatch, err
		}

		return data, http.MethodPatch, nil
	}

	// Regular object payload: use PUT.
	record, err := params.GetRecord()
	if err != nil {
		return nil, "", err
	}

	payloadWithCustomFields(record)

	if params.ObjectName == objectNameContacts {
		if err = c.payloadWithCommunicationItems(ctx, record); err != nil {
			return nil, "", err
		}

		if _, ok := record["communicationItems"]; ok {
			// Work around ConnectWise validation bug for contacts:
			// Including communicationItems in a PUT may cause a 400 even when
			// the items are valid. Since PUT is a full replacement, we clear
			// communicationItems first via PATCH, then proceed with the PUT.
			if err = c.clearContactCommunicationItems(ctx, params); err != nil {
				return nil, "", err
			}
		}
	}

	return record, http.MethodPut, nil
}

// clearContactCommunicationItems removes all communication items from a contact
// via a preliminary PATCH before a full PUT update.
//
// This is required due to a ConnectWise validation bug: when `communicationItems`
// are present in a PUT request for a contact, the API may reject the payload as
// invalid even if the items are correct. Clearing them first avoids this issue.
//
// The function performs a single PATCH with:
//   - "remove" operation for `/communicationItems`.
//   - "replace" operation for `/customFields` with an empty array to satisfy
//     ConnectWise's validation requirements for custom fields.
func (c *Connector) clearContactCommunicationItems(ctx context.Context,
	params common.WriteParams,
) error {
	url, err := c.getURL(objectNameContacts)
	if err != nil {
		return err
	}

	url.AddPath(params.RecordId)

	if _, err = c.JSONHTTPClient().Patch(ctx, url.String(), []map[string]any{
		{
			"op":   "remove",
			"path": "/communicationItems",
		},
		{
			"op":    "replace",
			"path":  "/customFields",
			"value": []any{},
		},
	}, c.clientIdHeader()); err != nil {
		return err
	}

	return nil
}

// payloadPatchWithCustomFields normalizes JSON Patch payloads for objects that
// support custom fields.
//
// ConnectWise requires that PATCH requests for such objects include a
// `/customFields` replace operation; otherwise the API may return 400 Bad
// Request, even if custom fields are not being modified.
//
// This function:
//   - Leaves payloads unchanged for objects that do not support custom fields.
//   - For supported objects:
//   - Extracts individual patch operations targeting custom field paths
//     (e.g. `/customField1`, `/customField2`) and converts them into entries
//     in a `/customFields` array of the form [{id, value}, ...].
//   - Preserves all non-custom-field operations as-is.
//   - Appends a single `/customFields` replace operation with the constructed
//     array.
//
// Notes:
//   - Only explicitly mentioned custom fields are included in the array.
//   - Omitting a custom field from the array does not clear its value.
//   - Sending an empty `customFields` array does not clear existing values.
func payloadPatchWithCustomFields(objectName string, payloads []patchOperationPayload) []patchOperationPayload {
	if !objectsSupportingCustomFields.Has(objectName) {
		// Object does not support custom fields; return payload unchanged.
		return payloads
	}

	customFields := make([]map[string]any, 0)
	result := make([]patchOperationPayload, 0)

	for _, payload := range payloads {
		// Normalize path by stripping leading slash.
		path, _ := strings.CutPrefix(payload.Path, "/")

		fieldIdStr, ok := strings.CutPrefix(path, "customField")
		if !ok {
			// Not a custom field operation; preserve as-is.
			result = append(result, payload)

			continue
		}

		fieldId, err := strconv.Atoi(fieldIdStr)
		if err != nil {
			// Malformed custom field ID; preserve operation as-is.
			result = append(result, payload)

			continue
		}

		customFields = append(customFields, map[string]any{
			"id":    fieldId,
			"value": payload.Value,
		})
	}

	// Always emit a /customFields replace operation for supported objects.
	result = append(result, patchOperationPayload{
		Op:    "replace",
		Path:  "/customFields",
		Value: customFields,
	})

	return result
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

// patchPayload represents a write payload that contains a JSON Patch array.
//
// The `patch` field holds a list of JSON Patch operations (add/replace/remove)
// targeting specific paths in the ConnectWise object. When this structure is
// detected, the connector uses HTTP PATCH and sends the inner array as the
// request body.
type patchPayload struct {
	Patch []patchOperationPayload `json:"patch"`
}

// patchOperationPayload represents a single JSON Patch operation as defined by RFC 6902.
type patchOperationPayload struct {
	Op          string `json:"op"`
	Path        string `json:"path"`
	Value       any    `json:"value,omitempty"`
	removeIndex int
}
