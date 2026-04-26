package odoo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var errEmptyID = errors.New("response returned empty id list")

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	body := map[string]any{}

	var method string

	if params.IsCreate() {
		method = "create"
		body["vals_list"] = []map[string]any{record}
	} else {
		method = "write"

		id, convErr := strconv.Atoi(strings.TrimSpace(params.RecordId))
		if convErr != nil {
			return nil, fmt.Errorf("record id must be numeric, got %q: %w", params.RecordId, convErr)
		}

		body["ids"] = []int{id}
		body["vals"] = record
	}

	urlStr, err := c.getURL(params.ObjectName, method)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal write body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create write request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	if params.IsUpdate() {
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
			Data:     nil,
			Errors:   nil,
		}, nil
	}

	recordID, err := odooCreateFirstID(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
	}, nil
}

// odooCreateFirstID reads the first id from Odoo create JSON (`[id, ...]`).
func odooCreateFirstID(body *ajson.Node) (string, error) {
	arr, err := jsonquery.New(body).ArrayRequired("")
	if err != nil {
		return "", fmt.Errorf("create response: %w", err)
	}

	if len(arr) == 0 {
		return "", errEmptyID
	}

	return jsonquery.New(arr[0]).TextWithDefault("", "")
}
