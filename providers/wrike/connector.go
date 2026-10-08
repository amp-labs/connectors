// Package wrike provides a connector for the Wrike API v4.
//
// API Documentation: https://developers.wrike.com/api/v4/
// Authentication: OAuth 2.0 authorization code (bearer token).
// Base URL: https://www.wrike.com/api (object paths carry the /v4 version prefix).
package wrike

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/wrike/metadata"
)

// Connector is the Wrike connector.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient

	components.SchemaProvider
}

// NewConnector creates a new Wrike connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Wrike, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
