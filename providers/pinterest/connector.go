package pinterest

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v5"

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	parameters.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
}

func NewConnector(params parameters.Connector) (*Connector, error) {
	// Create base connector with provider info
	return components.Initialize(providers.Pinterest, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeSerial,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	return connector, nil
}
