package salesforce

import (
	"context"
)

// BulkDelete launches async Bulk Job to delete records.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/create_job.htm
//
// After creation inspect newly launched Bulk Job via:
// * GetJobInfo
// * GetJobResults
// * GetSuccessfulJobResults.
func (c *Connector) BulkDelete(ctx context.Context, params BulkOperationParams) (*BulkOperationResult, error) {
	body := map[string]any{
		"object":    params.ObjectName,
		"operation": Delete,
	}

	return c.bulkOperation(ctx, params, body)
}
