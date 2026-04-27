// Package acculynx provides a connector for the AccuLynx V2 API.
//
// API Documentation: https://apidocs.acculynx.com/
// Authentication: API key sent as a Bearer token in the Authorization header.
// Base URL: https://api.acculynx.com
package acculynx

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/acculynx/metadata"
)

// Connector is the AccuLynx connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
}

// NewConnector creates a new AccuLynx connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.AccuLynx, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
