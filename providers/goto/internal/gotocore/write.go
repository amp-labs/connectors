package gotocore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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

	jsonData, err := marshalWriteBody(params.ObjectName, params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

// arrayBodyObjects are the GoTo objects whose write endpoints expect the request
// body to be a JSON array of records rather than a single object.
var arrayBodyObjects = datautils.NewSet("webhooks", "userSubscriptions") //nolint:gochecknoglobals

// marshalWriteBody serializes the write payload. RecordData always arrives as a
// single record object; most GoTo endpoints accept that object as-is, but a few
// (webhooks, userSubscriptions) require a JSON array, so for those we wrap the
// record in a one-element array before marshaling.
func marshalWriteBody(objectName string, recordData any) ([]byte, error) {
	asMap, err := common.RecordDataToMap(recordData)
	if err != nil {
		return nil, err
	}

	var payload any = asMap
	if arrayBodyObjects.Has(objectName) {
		payload = []map[string]any{asMap}
	}

	jsonData, err := json.Marshal(payload)
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

		// remove the "pages" suffix for corporate write endpoints for write operations
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
		recordID = extractWriteRecordID(data, cfg.writeIDFieldOrDefault())
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

	return extractGenericWriteRecord(body)
}

// extractWebinarWriteRecord pulls the first record out of `_embedded.<objectName>`,
// which is the envelope every Webinar write endpoint uses.
func extractWebinarWriteRecord(body *ajson.Node, objectName string) (map[string]any, string, error) {
	arr, err := jsonquery.New(body, "_embedded").ArrayOptional(objectName)
	if err != nil || len(arr) == 0 {
		return nil, "", fmt.Errorf("error extracting webinar write response: missing _embedded.%s", objectName) //nolint:err113,lll
	}

	data, err := jsonquery.Convertor.ObjectToMap(arr[0])
	if err != nil {
		return nil, "", fmt.Errorf("error converting webinar write response to map: %w", err)
	}

	return data, "", nil
}

// extractGenericWriteRecord reads the record from the response body. Most
// endpoints return a plain object; a few (e.g. meetings create) return a
// single-element array.
func extractGenericWriteRecord(body *ajson.Node) (map[string]any, string, error) {
	node := body
	if body.IsArray() {
		arr, err := body.GetArray()
		if err != nil || len(arr) == 0 {
			return nil, "", fmt.Errorf("empty array in write response: %w", err)
		}

		// for the objects that returns an array
		// It only returns a single record, so we take the first element of the array as the record node.
		node = arr[0]
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, "", fmt.Errorf("error converting write response to map: %w", err)
	}

	return data, "", nil
}

func extractWriteRecordID(data map[string]any, writeIDField string) string {
	switch v := data[writeIDField].(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	}

	return ""
}
