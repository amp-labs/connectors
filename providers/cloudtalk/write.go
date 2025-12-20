package cloudtalk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	readPath, err := c.getURLPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	baseResource := strings.TrimSuffix(readPath, "/index.json")

	var (
		url    *urlbuilder.URL
		method string
	)

	if params.RecordId != "" {
		// Update: POST /<resource>/edit/<id>.json
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, baseResource, "edit", params.RecordId+".json")
		if err != nil {
			return nil, err
		}

		method = http.MethodPost
	} else {
		// Create: PUT /<resource>/add.json
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, baseResource, "add.json")
		if err != nil {
			return nil, err
		}

		method = http.MethodPut
	}

	body, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	req *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	node, err := jsonquery.New(body).ObjectRequired("responseData")
	if err != nil {
		node = body
	}

	recordID := extractWriteID(node)

	if recordID == "" {
		recordID = params.RecordId
	}

	dataMap, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     dataMap,
	}, nil
}

func extractWriteID(node *ajson.Node) string {
	recordID, _ := jsonquery.New(node).TextWithDefault("id", "")
	if recordID != "" {
		return recordID
	}

	return extractNestedWriteID(node)
}

func extractNestedWriteID(node *ajson.Node) string {
	id, _ := jsonquery.New(node, "data").TextWithDefault("id", "")

	return id
}
