package asana

import (
	"context"
	"sync"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {

	var (
		wg sync.WaitGroup //nolint: varnamelen
		mu sync.Mutex     //nolint: varnamelen
	)

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.objectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

}
