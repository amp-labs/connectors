package schema

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components/operations"
)

// AggregateSchemaProvider gets the schema for multiple objects using a single batch request.
type AggregateSchemaProvider struct {
	operation *operations.ListObjectMetadataOperation
}

func NewAggregateSchemaProvider(
	client common.AuthenticatedHTTPClient,
	list operations.ListObjectMetadataHandlers,
) *AggregateSchemaProvider {
	return &AggregateSchemaProvider{
		operation: operations.NewHTTPOperation(client, list),
	}
}

func (p *AggregateSchemaProvider) ListObjectMetadata(
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

func (p *AggregateSchemaProvider) SchemaAcquisitionStrategy() string {
	return "AggregateSchemaProvider"
}
