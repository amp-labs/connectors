// Package okta provides a connector for the Okta Management API.
// API Documentation: https://developer.okta.com/docs/reference/core-okta-api/
package okta

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/okta/metadata"
)

// Connector is the Okta connector with metadata support.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
	components.SchemaProvider
}

// NewConnector creates a new Okta connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Okta, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Add SchemaProvider for metadata support
	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
