package parameters

import (
	"github.com/amp-labs/connectors/common"
)

// Connector is input parameters used to initialize a connector.
type Connector struct {
	Module common.ModuleID

	// AuthenticatedClient is a client for the connector that knows how to handle authentication for the provider.
	AuthenticatedClient common.AuthenticatedHTTPClient

	// Workspace is the provider workspace that the connector is connecting to. It could be a part of the metadata
	// map, but it is a common enough field that it is included here.
	Workspace string

	// Metadata is a map of key-value pairs that can be used to pass in additional information to the connector.
	// Generally this is used to substitute placeholders in the providerInfo, like workspace, server, etc, which is
	// information that is specific to the connection.
	Metadata map[string]string
}

// RequireModule must be embedded into connector struct to ensure that Module is supported.
type RequireModule struct {
	ExpectedModules []common.ModuleID
}

var _ moduleValidator = RequireModule{}

// RequireAuthenticatedClient must be embedded into connector struct to ensure that Client is set.
type RequireAuthenticatedClient struct{}

var _ authenticatedClientValidator = RequireAuthenticatedClient{}

// RequireWorkspace must be embedded into connector struct to ensure that Workspace is set.
type RequireWorkspace struct{}

var _ workspaceValidator = RequireWorkspace{}

// RequireMetadata must be embedded into connector struct to ensure that metadata keys are set.
// Keys are case-insensitive.
type RequireMetadata struct {
	ExpectedMetadataKeys []string
}

var _ metadataValidator = RequireMetadata{}
