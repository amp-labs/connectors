package breezy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// Breezy has no DELETE for positions. Archiving removes the position from the default
// published list (same practical outcome for integrations).
// https://developer.breezy.hr/reference/company-position-state-update
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if params.ObjectName != objectPositions {
		return nil, common.ErrOperationNotSupportedForObject
	}

	u, err := buildCompanyPositionStateURL(c.ProviderInfo().BaseURL, c.CompanyID, params.RecordId)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]string{"state": "archived"})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	_ context.Context,
	_ common.DeleteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if !httpkit.Status2xx(response.Code) {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	return &common.DeleteResult{Success: true}, nil
}
