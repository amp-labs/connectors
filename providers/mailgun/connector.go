// Package mailgun provides a connector for the Mailgun API.
//
// API Documentation: https://documentation.mailgun.com/docs/mailgun/api-reference/intro/
// Authentication: HTTP Basic auth ("api" as username, API key as password).
// Base URL: https://api.mailgun.net (US) or https://api.eu.mailgun.net (EU).
package mailgun

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// Connector is the Mailgun connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

// NewConnector creates a new Mailgun connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Mailgun, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
