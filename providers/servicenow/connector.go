package servicenow

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

const (
	restAPIPrefix = "api"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	// Require workspace
	common.RequireWorkspace
	// Require module
	common.RequireModule

	// Supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

func NewConnector(params common.Parameters) (*Connector, error) {
	return components.Initialize(providers.ServiceNow, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
		},
	)

	return connector, nil
}
