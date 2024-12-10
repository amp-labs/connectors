package metadata

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// OpenAPIStrategy is a ObjectMetadataStrategy that uses the OpenAPI schema to retrieve metadata for an object.
type OpenAPIStrategy struct {
	module common.Module
}

// This is sample code, but will need to be updated to actually read from the correct OpenAPI schema - seems like
// there is some relative pathing that needs to be done to get the correct schema.

func (i *OpenAPIStrategy) Run(ctx context.Context, object string) (*common.ObjectMetadata, error) {
	return metadata.Schemas.SelectOne(i.module.ID, object)
}
