package orb

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/orb/internal/metadata"
)

// ListObjectMetadata serves the catalog generated from Orb's OpenAPI spec. Orb has no describe
// endpoint, and the read objects have no custom fields beyond the free-form `metadata` map.
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Catalog.SelectResult(objectNames)
}
