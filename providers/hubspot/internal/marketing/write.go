package marketing

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	return nil, common.ErrNotImplemented
}

func (a *Adapter) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	return nil, common.ErrNotImplemented
}
