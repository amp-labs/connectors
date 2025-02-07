package operations

import (
	"github.com/amp-labs/connectors/common"
)

// Common operation types.
type (
	ReadHandlers               = HTTPHandlers[common.ReadParams, *common.ReadResult]
	WriteHandlers              = HTTPHandlers[common.WriteParams, *common.WriteResult]
	DeleteHandlers             = HTTPHandlers[common.DeleteParams, *common.DeleteResult]
	ListObjectMetadataHandlers = HTTPHandlers[[]string, *common.ListObjectMetadataResult]
)

// Common operation implementations.
type (
	ReadOperation               = HTTPOperation[common.ReadParams, *common.ReadResult]
	WriteOperation              = HTTPOperation[common.WriteParams, *common.WriteResult]
	DeleteOperation             = HTTPOperation[common.DeleteParams, *common.DeleteResult]
	ListObjectMetadataOperation = HTTPOperation[[]string, *common.ListObjectMetadataResult]
)
