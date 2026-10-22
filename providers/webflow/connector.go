// Package webflow provides a connector for the Webflow Data API v2.
//
// API Documentation: https://developers.webflow.com/data/reference
// Authentication: OAuth 2.0 authorization code (bearer token).
// Base URL: https://api.webflow.com (object paths carry the /v2 version prefix).
package webflow

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/webflow/metadata"
)

// metadataKeySiteID is the connection metadata key holding the Webflow site id.
const metadataKeySiteID = "siteId"

// Connector is the Webflow connector.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader
	components.Writer

	// SiteID scopes site-level object paths (/v2/sites/{site_id}/...).
	// Every object except `sites` requires it.
	SiteID string
}

// NewConnector creates a new Webflow connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Webflow, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector:            base,
		ExpectedMetadataKeys: []string{metadataKeySiteID},
		SiteID:               params.Metadata[metadataKeySiteID],
	}

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
