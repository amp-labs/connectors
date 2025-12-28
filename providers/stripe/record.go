package stripe

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// GetRecordsByIds fetches full records from Stripe for a specific set of IDs.
//
//nolint:revive
func (c *Connector) GetRecordsByIds(
	_ context.Context,
	_ string,
	_ []string,
	_ []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	return nil, fmt.Errorf("%w", errNotYetImplemented)
}
