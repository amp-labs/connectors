package justcall

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader
	components.Writer

	BaseURL  string
	ModuleID common.ModuleID
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.JustCall, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	providerInfo := connector.ProviderInfo()
	connector.BaseURL = providerInfo.BaseURL
	connector.ModuleID = connector.ProviderContext.Module()

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ModuleID, metadata.Schemas)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ModuleID,
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ModuleID,
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
