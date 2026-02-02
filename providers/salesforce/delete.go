package salesforce

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if c.crmAdapter != nil {
		return c.crmAdapter.Delete(ctx, params)
	}

	if c.pardotAdapter != nil {
		return c.pardotAdapter.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
