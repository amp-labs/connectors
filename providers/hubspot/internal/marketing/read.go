package marketing

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	return nil, common.ErrNotImplemented
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return nil, common.ErrNotImplemented
}
