package metadata

import (
	"context"
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// OpenAPIStrategy is a ObjectMetadataStrategy that uses the OpenAPI schema to retrieve metadata for an object.
type OpenAPIStrategy struct {
	module  common.ModuleID
	schemas *staticschema.Metadata
}

func NewOpenAPIStrategy(schemas *staticschema.Metadata, module common.ModuleID) *OpenAPIStrategy {
	return &OpenAPIStrategy{module: module, schemas: schemas}
}

func (i *OpenAPIStrategy) String() string {
	return "metadata.OpenAPIStrategy"
}

func (i *OpenAPIStrategy) GetObjectMetadata(
	_ context.Context,
	objects ...string,
) (*common.ListObjectMetadataResult, error) {
	return i.schemas.Select(i.module, objects)
}
