package salesforce

import (
	"context"
	"fmt"
)

// BulkWrite launches async Bulk Job to upsert records.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/create_job.htm
//
// After creation inspect newly launched Bulk Job via:
// * GetJobInfo
// * GetJobResults
// * GetSuccessfulJobResults.
func (c *Connector) BulkWrite( //nolint:funlen,cyclop
	ctx context.Context,
	params BulkOperationParams,
) (*BulkOperationResult, error) {
	// Only support upsert for now
	if params.Mode != Upsert {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMode, params.Mode)
	}

	body := map[string]any{
		"object":              params.ObjectName,
		"operation":           Upsert,
		"externalIdFieldName": params.ExternalIdField,
		"contentType":         "CSV",
		"lineEnding":          "LF",
	}

	result, err := c.bulkOperation(ctx, params, body)
	if err != nil {
		return nil, fmt.Errorf("bulk write failed: %w", err)
	}

	return result, nil
}
