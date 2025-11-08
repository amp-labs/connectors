package seismic

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	//go:embed schemas.json
	schemaContent []byte

	fileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV1](
		schemaContent, fileconv.NewSiblingFileLocator())

	schemas = fileManager.MustLoadSchemas()
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	// Supported operations
	components.SchemaProvider
	components.Reader
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	if params.Module == "" {
		params.Module = providers.ModuleSeismicReporting
	}
	// Create base connector with provider info
	return components.Initialize(providers.Seismic, params, constructor)
}

// nolint:funlen
func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	switch connector.Module() { //nolint: exhaustive
	case providers.ModuleSeismicReporting:
		// Set the metadata provider for the connector
		connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), schemas)

		// Set the read provider for the connector
		connector.Reader = reader.NewHTTPReader(
			connector.HTTPClient().Client,
			registry,
			connector.ProviderContext.Module(),
			operations.ReadHandlers{
				BuildRequest:  connector.buildReadRequest,
				ParseResponse: connector.parseReadResponse,
				ErrorHandler:  common.InterpretError,
			},
		)

	// We haven't implemented any other module upto now.
	default:
		return nil, common.ErrUnsupportedModule
	}

	return connector, nil
}
