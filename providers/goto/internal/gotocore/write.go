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
)

// recordIDCandidates lists the response keys we'll try, in order, when
// extracting the new record's id from a write response. GoTo's services
// don't agree on a single name (SCIM uses `id`, webinars use `webinarKey`,
// admin endpoints typically use `<entity>Key`).
var recordIDCandidates = []string{ //nolint:gochecknoglobals
	"id",
	"key",
	"webinarKey",
	"userKey",
	"groupKey",
}

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

		method = http.MethodPut
		if cfg.service == serviceSCIM {
			method = http.MethodPatch
		}
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
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
	data := map[string]any{}

	if _, ok := response.Body(); ok {
		if parsed, err := common.UnmarshalJSON[map[string]any](response); err == nil && parsed != nil {
			data = *parsed
		}
	}

	recordID := params.RecordId
	if recordID == "" {
		recordID = extractWriteRecordID(data)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func extractWriteRecordID(data map[string]any) string {
	for _, key := range recordIDCandidates {
		switch v := data[key].(type) {
		case string:
			if v != "" {
				return v
			}
		case float64:
			return fmt.Sprintf("%.0f", v)
		}
	}

	return ""
}
