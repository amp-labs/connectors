package braze

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Requires the instance id workspace
	common.RequireWorkspace

	// Supported operations
	components.SchemaProvider
}

const (
	limitQuery       = "limit"
	metadataPageSize = "1"
)

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// Create base connector with provider info
	return components.Initialize(providers.Braze, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.metadataRequest,
			ParseResponse: connector.parseMetadataResponse,
		},
	)

	return connector, nil
}
