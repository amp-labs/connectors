package connectwise

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise/internal/batch"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
	"github.com/amp-labs/connectors/providers/connectwise/internal/webhook"
)

const apiVersion = "v4_6_release/apis/3.0"

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
	// TODO must use webhook.Verifier instead of webhook.NoopVerifier
	*webhook.NoopVerifier

	batchAdapter *batch.Adapter // used for connectors.BatchRecordReaderConnector capabilities.

	clientID string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.ConnectWise, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	clientID := params.Metadata["clientId"]
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"clientId"},
		},
		clientID: clientID,
	}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle
	connector.SetErrorHandler(errorHandler)

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		common.ModuleRoot,
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.batchAdapter = batch.NewAdapter(connector.JSONHTTPClient(), connector.ProviderInfo(), clientID)
	connector.NoopVerifier = webhook.NewVerifier(connector.JSONHTTPClient(), connector.ProviderInfo(), clientID)

	return connector, nil
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ModuleInfo().BaseURL, apiVersion, objectPath)
}

func (c *Connector) clientIdHeader() common.Header {
	return common.Header{
		Key:   "ClientId",
		Value: c.clientID,
	}
}
