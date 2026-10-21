package orb

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/orb/internal/metadata"
)

// ListObjectMetadata serves the catalog generated from Orb's OpenAPI spec
// (scripts/openapi/orb). Orb has no describe endpoint; custom data lives in each
// record's free-form `metadata` object, which is listed as one field.
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Catalog.SelectResult(objectNames)
}
