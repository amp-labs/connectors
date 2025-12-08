package ads

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
	liinternal "github.com/amp-labs/connectors/providers/linkedin/internal/linkedininternal"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	//go:embed schemas.json
	schemaContent []byte

	fileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemaContent)

	schemas = fileManager.MustLoadSchemas()
)

type Adapter struct {
	*components.Connector
	common.RequireMetadata
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	AdAccountId string
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	conn, err := components.Initialize(providers.LinkedIn, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.AdAccountId = params.Metadata["adAccountId"]

	return conn, nil
}

// nolint:funlen
func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"adAccountId"},
		},
	}

	// LinkedIn's OpenAPI files only cover the adAnalytics object.
	// For other objects, we fall back to the sampling method to populate the fields.
	fallbackSchema := schema.NewObjectSchemaProvider(
		adapter.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  adapter.buildSingleObjectMetadataRequest,
			ParseResponse: adapter.parseSingleObjectMetadataResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(liinternal.ErrorFormats, nil),
			}.Handle,
		},
	)

	// The following method is specific to the 'adAnalytics' object.
	// See readme file for more info.
	adapter.SchemaProvider = schema.NewCompositeSchemaProvider(
		schema.NewOpenAPISchemaProvider(adapter.ProviderContext.Module(), schemas),
		fallbackSchema,
	)

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(liinternal.ErrorFormats, nil),
			}.Handle,
		},
	)

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(liinternal.ErrorFormats, nil),
			}.Handle,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(liinternal.ErrorFormats, nil),
			}.Handle,
		},
	)

	return adapter, nil
}
