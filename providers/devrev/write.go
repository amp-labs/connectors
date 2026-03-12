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
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// objectsWithFlatWriteResponse lists object names whose create/update API returns
// the record at the root (no wrapper key). Most DevRev write responses use a
// singular key, e.g. {"article": {...}}; auth-tokens.create returns a flat
// object {"access_token": "...", "expires_in": 3600, ...}.
var objectsWithFlatWriteResponse = datautils.NewSet( //nolint:gochecknoglobals
	"auth-tokens",
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// writeResponseKey derives the response object key from objectName.
// DevRev responses use singular snake_case: accounts->account, code-changes->code_change.
// The dot is a namespace separator (e.g. dev-orgs.auth-connections); only the last segment is used.
func writeResponseKey(objectName string) string {
	if idx := strings.LastIndex(objectName, "."); idx != -1 {
		objectName = objectName[idx+1:]
	}
	normalized := strings.ReplaceAll(objectName, "-", "_")

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

	var (
		recordNode *ajson.Node
		err        error
	)
	if objectsWithFlatWriteResponse.Has(params.ObjectName) {
		// Response is the record at root, e.g. auth-tokens: {"access_token": "...", "expires_in": 3600, ...}
		recordNode, err = jsonquery.New(body).ObjectRequired("")
	} else {
		// DevRev wraps the created/updated record under a singular key, e.g. {"article": {...}}
		responseKey := writeResponseKey(params.ObjectName)
		recordNode, err = jsonquery.New(body).ObjectRequired(responseKey)
	}

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
