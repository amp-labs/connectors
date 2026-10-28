// Package grain provides a connector for the Grain Public API.
//
// API Documentation: https://developers.grain.com/endpoints.html
// Authentication: OAuth2 authorization code with PKCE.
// Base URL: https://api.grain.com/_/public-api
package grain

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/grain/metadata"
)

// Connector is the Grain connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

// NewConnector creates a new Grain connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Grain, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
