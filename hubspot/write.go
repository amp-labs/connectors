package hubspot

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	return nil, fmt.Errorf("%w: Write", common.ErrNotImplemented)
}
