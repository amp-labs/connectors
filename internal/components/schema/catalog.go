package schema

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/objectcatalog"
)

var _ components.SchemaProvider = &CatalogSchemaProvider{}

// CatalogSchemaProvider implements SchemaProvider by serving object metadata from a static objectcatalog.Catalog.
type CatalogSchemaProvider struct {
	catalog objectcatalog.Catalog
}

func NewCatalogSchemaProvider(catalog objectcatalog.Catalog) *CatalogSchemaProvider {
	return &CatalogSchemaProvider{
		catalog: catalog,
	}
}

func (p *CatalogSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	return p.catalog.SelectResult(objects)
}

func (p *CatalogSchemaProvider) SchemaSource() string {
	return "CatalogSchemaProvider"
}
