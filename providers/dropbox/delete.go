package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// Delete is limited to closed file requests via POST /2/file_requests/delete.
// Other metadata objects are list-only or require routes not modeled here.
// https://docs.dropboxapi.com/dropbox-api/api-reference/user-endpoints/file-requests/delete
func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if params.ObjectName != objectFileRequests {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, fileRequestDeletePath)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]any{"ids": []string{params.RecordId}})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

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
