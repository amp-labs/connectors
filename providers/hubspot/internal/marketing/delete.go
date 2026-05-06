package marketing

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	return nil, common.ErrNotImplemented
}

func (a *Adapter) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	return nil, common.ErrNotImplemented
}
