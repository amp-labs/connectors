package grow

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/clio/internal/shared"
)

const clioGrowAPIPath = "grow"

func (c *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	return shared.BuildWriteRequest(ctx, shared.BuildWriteParams{
		BaseURL:     c.ProviderInfo().BaseURL,
		APIPath:     clioGrowAPIPath,
		Module:      c.Module(),
		Params:      params,
		FindURLPath: Schemas.FindURLPath,
	})
}

func (c *Adapter) parseWriteResponse(_ context.Context, params common.WriteParams,
	_ *http.Request, resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	return shared.ParseWriteResponse(params, resp)
}
