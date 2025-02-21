package schema

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components/operations"
)

// BatchSchemaProvider implements Provider using a single batch request
type BatchSchemaProvider struct {
	operation *operations.ListObjectMetadataOperation
}

func NewBatchSchemaProvider(
	client common.AuthenticatedHTTPClient,
	list operations.ListObjectMetadataHandlers,
) *BatchSchemaProvider {
	return &BatchSchemaProvider{
		operation: operations.NewHTTPOperation(client, list),
	}
}

func (p *BatchSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	if p.operation == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrNotImplemented, "schema provider is not implemented")
	}

	for _, object := range objects {
		if object == "" {
			return nil, fmt.Errorf("%w: object name cannot be empty", common.ErrMissingObjects)
		}
	}

	return p.operation.ExecuteRequest(ctx, objects)
}
