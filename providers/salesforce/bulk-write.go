package salesforce

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var ErrExternalIdEmpty = errors.New("external id is required")

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
	if len(params.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Only support upsert for now
	if params.Mode != UpsertMode {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMode, params.Mode)
	}

	if len(params.ExternalIdField) == 0 {
		// Upsert operation requires at least one field that is considered to be an External ID.
		// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/bulk_api_2_0_upsert.htm
		return nil, ErrExternalIdEmpty
	}

	if params.CSVData == nil {
		return nil, common.ErrMissingCSVData
	}

	body := map[string]any{
		"object":              params.ObjectName,
		"operation":           UpsertMode,
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
