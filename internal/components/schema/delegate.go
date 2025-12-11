package schema

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

var _ components.SchemaProvider = &DelegateSchemaProvider{}

// DelegateSchemaProvider implements SchemaProvider by delegating to a provided function.
// Use this when the connector needs full control over the schema implementation.
type DelegateSchemaProvider struct {
	execute func(context.Context, []string) (*common.ListObjectMetadataResult, error)
}

// NewDelegateSchemaProvider creates a new DelegateSchemaProvider with the provided list function.
func NewDelegateSchemaProvider(
	execute func(context.Context, []string) (*common.ListObjectMetadataResult, error),
) *DelegateSchemaProvider {
	return &DelegateSchemaProvider{
		execute: execute,
	}
}

func (p *DelegateSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objects) == 0 {
		return nil, common.ErrMissingObjects
	}

	return p.execute(ctx, objects)
}

func (p *DelegateSchemaProvider) SchemaSource() string {
	return "DelegateSchemaProvider"
}
