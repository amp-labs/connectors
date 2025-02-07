package schema

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// OpenAPIProvider implements Provider using OpenAPI schemas.
type OpenAPISchemaProvider[F staticschema.FieldMetadataMap] struct {
	module  common.ModuleID
	schemas *staticschema.Metadata[F]
}

func NewOpenAPISchemaProvider[F staticschema.FieldMetadataMap](
	module common.ModuleID,
	schemas *staticschema.Metadata[F],
) *OpenAPISchemaProvider[F] {
	return &OpenAPISchemaProvider[F]{
		module:  module,
		schemas: schemas,
	}
}

func (p *OpenAPISchemaProvider[F]) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	return p.schemas.Select(p.module, objects)
}
