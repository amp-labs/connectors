package gotocore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
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

	url, err := a.buildObjectBaseURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
		if cfg.service == serviceSCIM {
			method = http.MethodPut
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
