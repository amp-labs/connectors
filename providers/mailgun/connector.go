// Package mailgun provides a connector for the Mailgun API.
//
// API Documentation: https://documentation.mailgun.com/docs/mailgun/api-reference/intro/
// Authentication: HTTP Basic auth ("api" as username, API key as password).
// Base URL: https://api.mailgun.net (US) or https://api.eu.mailgun.net (EU).
package mailgun

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

const (
	regionEU  = "eu"
	euBaseURL = "https://api.eu.mailgun.net"
)

// Connector is the Mailgun connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader

	// workspace is the Mailgun sending domain (ConnectorParams.Workspace).
	// It is substituted into domain-scoped object paths at read time and is
	// only required for those objects (checked in buildReadRequest).
	workspace string
}

// NewConnector creates a new Mailgun connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Mailgun, params, constructor(params))
}

func constructor(params common.ConnectorParams) func(*components.Connector) (*Connector, error) {
	return func(base *components.Connector) (*Connector, error) {
		connector := &Connector{
			Connector: base,
			workspace: params.Workspace,
		}

		// Resolve the regional base URL. US (api.mailgun.net) is the catalog
		// default; EU is api.eu.mailgun.net. The US/EU host asymmetry prevents a
		// clean {{.region}} template, so the base URL is set here from metadata.
		if params.Metadata["region"] == regionEU {
			base.SetBaseURL(euBaseURL)
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

		return connector, nil
	}
}
