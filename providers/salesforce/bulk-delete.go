package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// BulkDelete launches async Bulk Job to delete records.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/create_job.htm
//
// After creation inspect newly launched Bulk Job via:
// * GetJobInfo
// * GetJobResults
// * GetSuccessfulJobResults.
func (c *Connector) BulkDelete(ctx context.Context, params BulkOperationParams) (*BulkOperationResult, error) {
	if len(params.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if params.CSVData == nil {
		return nil, common.ErrMissingCSVData
	}

	body := map[string]any{
		"object":    params.ObjectName,
		"operation": DeleteMode,
	}

	return c.bulkOperation(ctx, params, body)
}
