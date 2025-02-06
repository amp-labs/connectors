package schema

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components/operations"
)

// HTTPObjectSchemaProvider implements Provider using HTTP.
type HTTPObjectSchemaProvider struct {
	operation *operations.ListObjectMetadataOperation
}

func NewHTTPObjectSchemaProvider(
	client common.AuthenticatedHTTPClient,
	list operations.ListObjectMetadataHandlers,
) *HTTPObjectSchemaProvider {
	return &HTTPObjectSchemaProvider{
		operation: operations.NewHTTPOperation(client, list),
	}
}

func (p *HTTPObjectSchemaProvider) GetMetadata(
	ctx context.Context,
	objects ...string,
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
