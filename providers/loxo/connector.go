package loxo

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
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
	common.RequireMetadata

	// Supported operations
	components.SchemaProvider

	AgencySlug string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// Create base connector with provider info
	conn, err := components.Initialize(providers.Loxo, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.AgencySlug = params.Metadata["agencySlug"]

	return conn, nil
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
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	return connector, nil
}
