package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if c.isPardotModule() {
		return c.pardotAdapter.Delete(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
