// Package gusto provides a connector for the Gusto HR & Payroll API.
// API Documentation: https://docs.gusto.com/app-integrations/reference
// Authentication: OAuth 2.0 Authorization Code
// Base URL: https://api.gusto.com
package gusto

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
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
	common.PostAuthInfo

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	companyId string
}

// NewConnector creates a new Gusto connector for the production environment.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Gusto, params, constructor)
}

// NewDemoConnector creates a new Gusto connector for the sandbox/demo environment.
func NewDemoConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.GustoDemo, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	authMetadata := NewAuthMetadataVars(params.Metadata)

	connector := &Connector{
		Connector: base,
		companyId: authMetadata.CompanyId,
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

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
