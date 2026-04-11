// Package gusto provides a connector for the Gusto HR & Payroll API.
// API Documentation: https://docs.gusto.com/app-integrations/reference
// Authentication: OAuth 2.0 Authorization Code
// Base URL: https://api.gusto.com
package gusto

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gusto/metadata"
)

// Connector is the Gusto connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

// NewConnector creates a new Gusto connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Gusto, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
