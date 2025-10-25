package linkedin

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
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

	fileManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV2](
		schemaContent, fileconv.NewSiblingFileLocator())

	schemas = fileManager.MustLoadSchemas()
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireMetadata

	// Supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	AdAccountId string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// Create base connector with provider info
	conn, err := components.Initialize(providers.LinkedIn, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.AdAccountId = params.Metadata["adAccountId"]

	return conn, nil
}

//nolint:funlen
func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// LinkedIn's OpenAPI files only cover the adAnalytics object.
	// For other objects, we fall back to the sampling method to populate the fields.
	fallbackSchema := schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	// The following method is specific to the 'adAnalytics' object.
	// See readme file for more info.
	connector.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), schemas),
		fallbackSchema,
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	return connector, nil
}
