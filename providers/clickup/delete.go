package clickup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	spec, ok := writeObjects[params.ObjectName]
	if !ok {
		return nil, common.ErrObjectNotSupported
	}

	if spec.deletePath == "" {
		return nil, fmt.Errorf("%w: object %v", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	url, err := c.getWriteDeleteURL(spec.deletePath, params.RecordId)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
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

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
