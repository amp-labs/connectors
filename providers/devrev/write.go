package devrev

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// DevRev uses POST for both create and update, with object.create and object.update endpoints.
// Example: POST /accounts.create, POST /accounts.update (id in body).
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	objectName := params.ObjectName
	if params.RecordId != "" {
		objectName += ".update"
	} else {
		objectName += ".create"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	if err != nil {
		return nil, err
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		recordData["id"] = params.RecordId
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
}

// writeResponseKey derives the response object key from objectName.
// DevRev responses use singular snake_case: accounts->account, code-changes->code_change.
func writeResponseKey(objectName string) string {
	normalized := strings.ReplaceAll(objectName, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")

	return naming.NewSingularString(normalized).String()
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	// DevRev wraps the created/updated record under a singular key, e.g. {"article": {...}}
	responseKey := writeResponseKey(params.ObjectName)

	recordNode, err := jsonquery.New(body).ObjectRequired(responseKey)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(recordNode).StrWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(recordNode)
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
