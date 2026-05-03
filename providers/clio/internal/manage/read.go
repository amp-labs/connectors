package manage

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/clio/internal/shared"
)

const (
	clioManageAPIPath = "api/v4"
)

// ObjectsNoUpdatedSince is a set of objects that do not support the provider-side `updated_since`.
var ObjectsNoUpdatedSince = datautils.NewSet( //nolint:gochecknoglobals
	"lauk_civil_certificated_rates",
	"lauk_civil_controlled_rates",
	"lauk_criminal_controlled_rates",
	"clio_payments/payments",
)

func (c *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	return shared.BuildReadRequest(ctx, shared.BuildReadParams{
		BaseURL:               c.ProviderInfo().BaseURL,
		APIPath:               clioManageAPIPath,
		Module:                c.Module(),
		Params:                params,
		FindURLPath:           Schemas.FindURLPath,
		ObjectsNoUpdatedSince: ObjectsNoUpdatedSince,
	})
}

func (c *Adapter) parseReadResponse(_ context.Context, params common.ReadParams,
	_ *http.Request, resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return shared.ParseReadResponse(params, resp, c.Module(), Schemas.LookupArrayFieldName, ObjectsNoUpdatedSince)
}
