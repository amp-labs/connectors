package odoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// odooUnlink is the JSON body for POST .../{model}/unlink.
type odooUnlink struct {
	IDs []int `json:"ids"`
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	id, err := strconv.Atoi(strings.TrimSpace(params.RecordId))
	if err != nil {
		return nil, fmt.Errorf("odoo: record id must be numeric, got %q: %w", params.RecordId, err)
	}

	urlStr, err := c.getURL(params.ObjectName, "unlink")
	if err != nil {
		return nil, err
	}

	payload := odooUnlink{IDs: []int{id}}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal odoo unlink body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create odoo unlink request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	_ *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	// Odoo unlink returns true on success; failures use error HTTP.
	return &common.DeleteResult{Success: true}, nil
}
