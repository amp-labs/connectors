package monaco

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/monaco/metadata"
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	components.SchemaProvider
	components.Reader
	components.Writer
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Monaco, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Fields come from the static schemas.json generated off Monaco's public
	// OpenAPI spec. Monaco also serves GET /v1/schemas/{entity} at runtime,
	// which additionally reports per-org custom fields (custom_field_<uuid>)
	// and each field's allowed_operators, but it only covers 6 of the 11
	// objects. Layering that in is deferred until we have credentials to
	// verify its payload shape.
	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.Module(),
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
