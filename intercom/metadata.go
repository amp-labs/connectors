package intercom

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/intercom/metadata"
)

var ErrLoadFailure = errors.New("cannot load metadata")

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	schemas, err := metadata.LoadSchemas()
	if err != nil {
		return nil, ErrLoadFailure
	}

	return schemas.Select(objectNames)
}
