package pardot

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

func (a *Adapter) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	objectNameLower := strings.ToLower(params.ObjectName)

	url, err := a.getURL(objectNameLower)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	common.Headers(common.TransformWriteHeaders(params.Headers, common.HeaderModeOverwrite)).ApplyToRequest(req)

	req.Header.Set("Pardot-Business-Unit-Id", a.businessUnitID)

	return req, nil
}

func (a *Adapter) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if !httpkit.Status2xx(response.Code) {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{
		Success: true,
	}, nil
}
