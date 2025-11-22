package bitbucket

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
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
	components.Writer

	Workspace string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Bitbucket, params, constructor)
	if err != nil {
		return nil, err
	}

	connector.Workspace = params.Workspace

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Bitbucket's OpenAPI files don't cover all API resources,
	// so we fall back to querying objects and populating fields.
	fallbackSchema := schema.NewObjectSchemaProvider(
		base.HTTPClient().Client,
		schema.FetchModeSerial,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleHandlerRequest,
			ParseResponse: connector.parseSingleHandlerResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(common.ModuleRoot, schemas),
		fallbackSchema,
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		common.ModuleRoot,
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
