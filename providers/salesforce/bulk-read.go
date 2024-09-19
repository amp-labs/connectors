package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// BulkRead launches async Query job for bulk reading.
// It is similar to BulkQuery. Check it for more info.
//
// Usage example:
//
//	BulkRead(ctx, common.ReadParams{
//		ObjectName: "accounts",
//		Since: time.Now().Add(-15 * time.Minute)
//		Deleted: true,
//	})
func (c *Connector) BulkRead(ctx context.Context, params common.ReadParams) (*GetJobInfoResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	soql := makeSOQL(params)
	// Note: if params.Deleted is set to true query will return only removed items.

	query := soql.String()

	return c.BulkQuery(ctx, query, params.Deleted)
}
