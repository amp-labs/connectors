package pipeliner

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/pipeliner/metadata"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	schemas, err := metadata.FileManager.LoadSchemas()
	if err != nil {
		return nil, common.ErrMetadataLoadFailure
	}

	return schemas.Select(objectNames)
}
