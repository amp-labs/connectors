package components

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Reader represents the ability to read objects from a data source.
type Reader interface {
	// Read retrieves an object from the data source
	Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
}

// Writer represents the ability to write objects to a data source.
type Writer interface {
	Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)
}

// Deleter represents the ability to delete objects from a data source.
type Deleter interface {
	Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error)
}

// SchemaProvider represents the ability to retrieve metadata about objects.
type SchemaProvider interface {
	ListObjectMetadata(ctx context.Context, objects ...string) (*common.ListObjectMetadataResult, error)
}
