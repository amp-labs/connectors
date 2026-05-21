// Package gusto provides a connector for the Gusto HR & Payroll API.
// API Documentation: https://docs.gusto.com/app-integrations/reference
// Authentication: OAuth 2.0 Authorization Code
// Base URL: https://api.gusto.com
package gusto

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gusto/metadata"
)

// metadataKeyCompanyID is the key under ConnectorParams.Metadata that must
// contain the Gusto company UUID the installation is scoped to.
const metadataKeyCompanyID = "companyId"

// Connector is the Gusto connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader

	companyID string
}

// NewConnector creates a new Gusto connector for the production environment.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Gusto, params, constructor(params))
}

// NewDemoConnector creates a new Gusto connector for the sandbox/demo environment.
func NewDemoConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.GustoDemo, params, constructor(params))
}

func constructor(params common.ConnectorParams) func(*components.Connector) (*Connector, error) {
	return func(base *components.Connector) (*Connector, error) {
		connector := &Connector{
			Connector: base,
			companyID: params.Metadata[metadataKeyCompanyID],
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
