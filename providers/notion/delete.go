package notion

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// Notion does not delete users or file uploads.
// https://developers.notion.com/reference/get-users
// https://developers.notion.com/reference/list-file-uploads
//
// Delete is rejected in supportedOperations() (no DeleteSupport); HTTPDeleter checks the
// registry first, so the handlers below are not reached for catalog objects today.
func (c *Connector) buildDeleteRequest(_ context.Context, params common.DeleteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return nil, common.ErrOperationNotSupportedForObject
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
