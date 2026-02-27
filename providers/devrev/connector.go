package devrev

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/devrev/metadata"
)

type Connector struct {
	*components.Connector
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
	common.RequireAuthenticatedClient
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.DevRev, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas),
	)
	registry := components.NewEmptyEndpointRegistry()

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
	// connector.Deleter = deleter.NewHTTPDeleter(
	// 	connector.HTTPClient().Client,
	// 	registry,
	// 	connector.ProviderContext.Module(),
	// 	operations.DeleteHandlers{
	// 		BuildRequest:  connector.buildDeleteRequest,
	// 		ParseResponse: connector.parseDeleteResponse,
	// 		ErrorHandler:  common.InterpretError,
	// 	},
	// )
	return connector, nil
}
