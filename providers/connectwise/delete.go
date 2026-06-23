package connectwise

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	c.clientIdHeader().ApplyToRequest(req)

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if !httpkit.Status2xx(resp.Code) {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
