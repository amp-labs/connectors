package connectors

import (
	"context"
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Connector is an interface that can be used to implement a connector with
// basic configuration about the provider.
type Connector interface {
	fmt.Stringer
	io.Closer

	// JSONHTTPClient returns the underlying JSON HTTP client. This is useful for
	// testing, or for calling methods that aren't exposed by the Connector
	// interface directly. Authentication and token refreshes will be handled automatically.
	JSONHTTPClient() *common.JSONHTTPClient

	// HTTPClient returns the underlying HTTP client. This is useful for proxy requests.
	HTTPClient() *common.HTTPClient

	// Provider returns the connector provider.
	Provider() providers.Provider
}

// ReadConnector is an interface that extends the Connector interface with read capabilities.
type ReadConnector interface {
	Connector

	// Read reads a page of data from the connector. This can be called multiple
	// times to read all the data. The caller is responsible for paging, by
	// passing the NextPage value correctly, and by terminating the loop when
	// Done is true. The caller is also responsible for handling errors.
	// Authentication corner cases are handled internally, but all other errors
	// are returned to the caller.
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)
}

// WriteConnector is an interface that extends the Connector interface with write capabilities.
type WriteConnector interface {
	Connector

	Write(ctx context.Context, params WriteParams) (*WriteResult, error)
}

// DeleteConnector is an interface that extends the Connector interface with delete capabilities.
type DeleteConnector interface {
	Connector

	Delete(ctx context.Context, params DeleteParams) (*DeleteResult, error)
}

// ObjectMetadataConnector is an interface that extends the Connector interface with
// the ability to list object metadata.
type ObjectMetadataConnector interface {
	Connector

	ListObjectMetadata(ctx context.Context, objectNames []string) (*ListObjectMetadataResult, error)
}

// AuthMetadataConnector is an interface that extends the Connector interface with
// the ability to retrieve metadata information about authentication.
type AuthMetadataConnector interface {
	Connector

	// GetPostAuthInfo returns authentication metadata.
	GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error)
}

// We re-export the following types so that they can be used by consumers of this library.
type (
	ReadParams               = common.ReadParams
	WriteParams              = common.WriteParams
	DeleteParams             = common.DeleteParams
	ReadResult               = common.ReadResult
	WriteResult              = common.WriteResult
	DeleteResult             = common.DeleteResult
	ListObjectMetadataResult = common.ListObjectMetadataResult

	ErrorWithStatus = common.HTTPStatusError
)
