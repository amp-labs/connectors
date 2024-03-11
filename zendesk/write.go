package zendesk

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// TODO:: Implement me
// Temporarily added as empty func to satisfy interface method requirements

// Write will write data to Zendesk.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	return &common.WriteResult{}, nil
}
