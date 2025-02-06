package schema

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// OpenAPIProvider implements Provider using OpenAPI schemas.
type OpenAPISchemaProvider struct {
	module  common.ModuleID
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV1]
}

func NewOpenAPISchemaProvider(
	module common.ModuleID,
	schemas *staticschema.Metadata[staticschema.FieldMetadataMapV1],
) *OpenAPISchemaProvider {
	return &OpenAPISchemaProvider{
		module:  module,
		schemas: schemas,
	}
}

func (p *OpenAPISchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects ...string,
) (*common.ListObjectMetadataResult, error) {
	return p.schemas.Select(p.module, objects)
}
