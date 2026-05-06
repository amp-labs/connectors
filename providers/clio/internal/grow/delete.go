package grow

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/clio/internal/shared"
)

func (c *Adapter) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	return shared.BuildDeleteRequest(ctx, shared.BuildDeleteParams{
		BaseURL:     c.ProviderInfo().BaseURL,
		APIPath:     clioGrowAPIPath,
		Module:      c.Module(),
		Params:      params,
		FindURLPath: Schemas.FindURLPath,
	})
}

func (c *Adapter) parseDeleteResponse(_ context.Context, params common.DeleteParams,
	_ *http.Request, resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	return shared.ParseDeleteResponse(params, resp)
}
