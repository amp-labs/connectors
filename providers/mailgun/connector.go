// Package mailgun provides a connector for the Mailgun API.
//
// API Documentation: https://documentation.mailgun.com/docs/mailgun/api-reference/intro/
// Authentication: HTTP Basic auth ("api" as username, API key as password).
// Base URL: https://api.mailgun.net (US) or https://api.eu.mailgun.net (EU).
package mailgun

import (
	"strings"

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

	// workspace is the Mailgun sending domain (ConnectorParams.Workspace),
	// substituted into domain-scoped object paths and required only for those
	// objects (enforced in substituteDomain).
	workspace string
}

// NewConnector creates a new Mailgun connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Mailgun, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		workspace: params.Workspace,
	}

	// US (api.mailgun.net) is the catalog default base URL; only EU needs an
	// override. Region is free-text metadata, so match it case-insensitively.
	if strings.EqualFold(strings.TrimSpace(params.Metadata["region"]), regionEU) {
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
