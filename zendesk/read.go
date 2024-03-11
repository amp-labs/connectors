package zendesk

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// TODO:: Implement me
// Temporarily added as empty func to satisfy interface method requirements

// Read reads data from Zendesk.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	return &common.ReadResult{}, nil
}
