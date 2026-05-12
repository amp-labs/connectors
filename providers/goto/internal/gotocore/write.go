package gotocore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	cfg, ok := objectRegistry[params.ObjectName]
	if !ok || !cfg.writable {
		return nil, fmt.Errorf("%w: object %s does not support write",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	url, err := a.buildWriteObjectURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		// Most GoTo update endpoints use PUT; Webinar uses PATCH for partial
		// updates.
		method = http.MethodPut
		if cfg.service == serviceWebinar || cfg.service == serviceSCIM {
			method = http.MethodPatch
		}
	}

	jsonData, err := marshalWriteBody(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

// marshalWriteBody serializes the write payload. Most GoTo endpoints accept a
// JSON object, but a few (e.g. webhooks, userSubscriptions) accept an array of
// objects — for those we pass the slice through without coercing it to a map.
func marshalWriteBody(recordData any) ([]byte, error) {
	v := reflect.ValueOf(recordData)
	if v.IsValid() && (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) {
		jsonData, err := json.Marshal(recordData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal record data: %w", err)
		}

		return jsonData, nil
	}

	asMap, err := common.RecordDataToMap(recordData)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(asMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	return jsonData, nil
}

// buildWriteObjectURL resolves the object path against the module BaseURL,
// substituting the account key. Callers attach their own query params.
func (a *Adapter) buildWriteObjectURL(objectName string) (*urlbuilder.URL, error) {
	spec, ok := objectRegistry[objectName]
	if !ok || spec.path == "" {
		spec.path = objectName
	}

	if spec.service == serviceCorporate {
		path := strings.ReplaceAll(spec.path, accountKeyPlaceholder, a.accountKey)

		//remove the "pages" suffix for corporate write endpoints for write operations
		path = strings.TrimSuffix(path, "/pages")

		url, err := urlbuilder.New(a.ModuleInfo().BaseURL, path)
		if err != nil {
			return nil, fmt.Errorf("error building URL for object %s: %w", objectName, err)
		}
		return url, nil
	}

	path := strings.ReplaceAll(spec.path, accountKeyPlaceholder, a.accountKey)

	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, path)
	if err != nil {
		return nil, fmt.Errorf("error building URL for object %s: %w", objectName, err)
	}

	return url, nil
}

func (a *Adapter) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	cfg := objectRegistry[params.ObjectName]

	data, recordID, err := extractWriteResponse(body, cfg, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		recordID = params.RecordId
	} else if recordID == "" {
		recordID = extractWriteRecordID(data)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

// extractWriteResponse normalizes the various shapes GoTo write endpoints
// return. It returns the record map, optionally a record ID (when the body is
// just a bare string), and an error.
func extractWriteResponse(body *ajson.Node, cfg objectConfig, objectName string) (map[string]any, string, error) {
	// Some admin endpoints (e.g. attributes) return the new record's id as a
	// bare JSON string with no surrounding object.
	if body.IsString() {
		id, err := body.GetString()
		if err != nil {
			return nil, "", fmt.Errorf("error reading bare-string write response: %w", err)
		}

		return map[string]any{}, id, nil
	}

	if cfg.service == serviceWebinar {
		return extractWebinarWriteRecord(body, objectName)
	}

	return extractGenericWriteRecord(body, "")
}

// extractWebinarWriteRecord pulls the first record out of `_embedded.<objectName>`,
// which is the envelope every Webinar write endpoint uses.
func extractWebinarWriteRecord(body *ajson.Node, objectName string) (map[string]any, string, error) {
	arr, err := jsonquery.New(body, "_embedded").ArrayOptional(objectName)
	if err != nil || len(arr) == 0 {
		return nil, "", fmt.Errorf("error extracting webinar write response: missing _embedded.%s", objectName)
	}

	data, err := jsonquery.Convertor.ObjectToMap(arr[0])
	if err != nil {
		return nil, "", fmt.Errorf("error converting webinar write response to map: %w", err)
	}

	return data, "", nil
}

// extractGenericWriteRecord reads the record from the configured writeResponseKey,
// falling back to a single-element array when present (e.g. meetings create).
func extractGenericWriteRecord(body *ajson.Node, responseKey string) (map[string]any, string, error) {
	resp, err := jsonquery.New(body).ObjectOptional(responseKey)
	if err != nil || resp == nil {
		arr, arrErr := jsonquery.New(body).ArrayOptional(responseKey)
		if arrErr != nil || len(arr) == 0 {
			return nil, "", fmt.Errorf("error extracting write response data: %w", err)
		}

		resp = arr[0]
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, "", fmt.Errorf("error converting write response to map: %w", err)
	}

	return data, "", nil
}

func extractWriteRecordID(data map[string]any) string {
	switch v := data["id"].(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	}

	return ""
}
