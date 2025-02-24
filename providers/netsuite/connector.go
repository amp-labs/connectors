package netsuite

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	// Supported operations
	components.SchemaProvider
}

func NewConnector(params common.Parameters) (*Connector, error) {
	// Create base connector with provider info
	return components.Initialize(providers.Netsuite, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildObjectMetadataRequest,
			ParseResponse: connector.parseObjectMetadataResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
