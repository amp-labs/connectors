// Package wrike provides a connector for the Wrike API v4.
//
// API Documentation: https://developers.wrike.com/api/v4/
// Authentication: OAuth 2.0 authorization code (bearer token).
// Base URL: https://www.wrike.com/api (object paths carry the /v4 version prefix).
package wrike

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/wrike/metadata"
)

// Connector is the Wrike connector.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader
	components.Writer
}

// NewConnector creates a new Wrike connector for accounts on www.wrike.com.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return NewConnectorForProvider(providers.Wrike, params)
}

// NewConnectorForProvider creates a Wrike connector under a regional provider
// (Wrike, WrikeEU, WrikeUS2). The implementation is shared; only the catalog
// base URL, i.e. the data centre host, differs.
func NewConnectorForProvider(provider providers.Provider, params common.ConnectorParams) (*Connector, error) {
	return components.Init(provider, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
