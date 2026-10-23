package sybill

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/sybill/internal/metadata"
)

// ListObjectMetadata serves the catalog generated from the OpenAPI spec. Sybill has no describe
// endpoint, and its objects carry standard fields only. Conversations include the detail-only
// fields (transcript, summary, recordings, category) that Read fetches per conversation.
func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return metadata.Catalog.SelectResult(objectNames)
}
