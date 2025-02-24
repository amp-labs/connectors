package operations

import (
	"github.com/amp-labs/connectors/common"
)

// Common operation types.
type (
	ReadHandlers   = HTTPHandlers[common.ReadParams, *common.ReadResult]
	WriteHandlers  = HTTPHandlers[common.WriteParams, *common.WriteResult]
	DeleteHandlers = HTTPHandlers[common.DeleteParams, *common.DeleteResult]

	// Gets metadata for a list of objects in a single request.
	ListObjectMetadataHandlers = HTTPHandlers[[]string, *common.ListObjectMetadataResult]

	// Gets metadata for a single object.
	SingleObjectMetadataHandlers = HTTPHandlers[string, *common.ObjectMetadata]
)

// Common operation implementations.
type (
	ReadOperation   = HTTPOperation[common.ReadParams, *common.ReadResult]
	WriteOperation  = HTTPOperation[common.WriteParams, *common.WriteResult]
	DeleteOperation = HTTPOperation[common.DeleteParams, *common.DeleteResult]

	// Gets metadata for a list of objects in a single request.
	ListObjectMetadataOperation = HTTPOperation[[]string, *common.ListObjectMetadataResult]

	// Gets metadata for a single object.
	SingleObjectMetadataOperation = HTTPOperation[string, *common.ObjectMetadata]
)
