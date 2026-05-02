package grow

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/clio/internal/shared"
)

const (
	clioGrowAPIPath = "grow" // https://api.clio.com/grow/... (Clio Platform / Grow)
)

func (c *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	return shared.BuildReadRequest(ctx, shared.BuildReadParams{
		BaseURL:               c.ProviderInfo().BaseURL,
		APIPath:               clioGrowAPIPath,
		Module:                c.Module(),
		Params:                params,
		FindURLPath:           Schemas.FindURLPath,
		ObjectsNoUpdatedSince: nil,
	})
}

func (c *Adapter) parseReadResponse(_ context.Context, params common.ReadParams,
	_ *http.Request, resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return shared.ParseReadResponse(params, resp, c.Module(), Schemas.LookupArrayFieldName, nil)
}
