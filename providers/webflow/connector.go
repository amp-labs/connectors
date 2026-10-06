// Package webflow provides a connector for the Webflow Data API v2.
//
// API Documentation: https://developers.webflow.com/data/reference
// Authentication: OAuth 2.0 authorization code (bearer token).
// Base URL: https://api.webflow.com (object paths carry the /v2 version prefix).
package webflow

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/webflow/metadata"
)

// Connector is the Webflow connector.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	components.SchemaProvider
}

// NewConnector creates a new Webflow connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Webflow, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
