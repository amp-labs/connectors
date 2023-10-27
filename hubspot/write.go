package hubspot

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	return nil, common.ErrNotImplemented
}
