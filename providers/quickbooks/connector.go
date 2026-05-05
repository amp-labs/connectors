package quickbooks

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

const (
	restAPIPrefix = "v3/company"
	pageSize      = "1000"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireMetadata

	// Supported operations
	components.SchemaProvider

	components.Reader
	components.Writer

	// Workspace is the Company ID (realmId) in QuickBooks.
	// http://developer.intuit.com/app/developer/qbo/docs/develop/authentication-and-authorization/oauth-2.0
	Workspace string
	// graphQLBaseURL is a variable on the struct so it can be mocked in unit tests.
	graphQLBaseURL string
}

// resolveWorkspace returns the QuickBooks workspace ref (realmId), preferring
// params.Workspace and falling back to params.Metadata["realmId"] for legacy
// connections that stored the realmId in metadata.
func resolveWorkspace(params common.ConnectorParams) string {
	if params.Workspace != "" {
		return params.Workspace
	}

	return params.Metadata["realmId"]
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Initialize(providers.QuickBooks, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.Workspace = resolveWorkspace(params)

	return conn, nil
}

func NewSandboxConnector(params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Initialize(providers.QuickbooksSandbox, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.Workspace = resolveWorkspace(params)

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeSerial,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
