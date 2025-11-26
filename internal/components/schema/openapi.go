package schema

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// OpenAPISchemaProvider implements Provider using OpenAPI schemas.
type OpenAPISchemaProvider[F staticschema.FieldMetadataMap, C any] struct {
	module  common.ModuleID
	schemas *staticschema.Metadata[F, C]
}

func NewOpenAPISchemaProvider[F staticschema.FieldMetadataMap, C any](
	module common.ModuleID,
	schemas *staticschema.Metadata[F, C],
) *OpenAPISchemaProvider[F, C] {
	return &OpenAPISchemaProvider[F, C]{
		module:  module,
		schemas: schemas,
	}
}

func (p *OpenAPISchemaProvider[F, C]) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	return p.schemas.Select(p.module, objects)
}

func (p *OpenAPISchemaProvider[F, C]) SchemaAcquisitionStrategy() string {
	return "OpenAPISchemaProvider"
}
