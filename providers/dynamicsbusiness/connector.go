package dynamicsbusiness

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireWorkspace
	common.RequireMetadata

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	environmentName string
	tenantID        string
	companyID       string

	incrementalRegistry *datautils.Cache[string, bool]
}

const (
	metadataKeyCompanyID       = "companyId"
	metadataKeyEnvironmentName = "environmentName"
)

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Initialize(providers.DynamicsBusinessCentral, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.tenantID = params.Workspace
	conn.companyID = params.Metadata[metadataKeyCompanyID]
	conn.environmentName = params.Metadata[metadataKeyEnvironmentName]

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{metadataKeyCompanyID, metadataKeyEnvironmentName},
		},
		incrementalRegistry: datautils.NewCache[string, bool](),
	}

	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
		},
	)

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		XML:  interpreter.NewXMLFaultyResponder(xmlErrorFormats, nil),
	}.Handle

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
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

	return connector, nil
}
