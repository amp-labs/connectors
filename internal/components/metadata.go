package components

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// MetadataStrategy describes an object's schema / metadata.
type MetadataStrategy interface {
	// Run accepts an object, and returns a field map for it.
	Run(ctx context.Context, object string) (*common.ObjectMetadata, error)
}

type objectMetadataResult struct {
	ObjectName string
	Response   common.ObjectMetadata
}

type objectMetadataError struct {
	ObjectName string
	Error      error
}

func (c *ConnectorComponent) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {
	if _, ok := c.MetadataStrategy.(MetadataStrategy); !ok {
		return nil, common.ErrNotImplemented
	}

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	objectsMap := &common.ListObjectMetadataResult{}
	objectsMap.Result = make(map[string]common.ObjectMetadata)
	objectsMap.Errors = make(map[string]error)

	for _, object := range objectNames {
		response, err := c.MetadataStrategy.Run(ctx, object)
		if err != nil {
			objectsMap.Errors[object] = err
		} else {
			objectsMap.Result[object] = *response
		}
	}

	return objectsMap, nil
}
