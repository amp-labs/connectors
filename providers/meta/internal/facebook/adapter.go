package facebook

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	// Basic connector
	*components.Connector

	common.RequireMetadata
	// Supported operations
	components.SchemaProvider
	components.Reader

	adAccountId string
	businessId  string
}

const (
	metadataKeyAdAccountID = "adAccountId"
	metadataKeyBusinessID  = "businessId"
)

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	// Create base connector with provider info
	conn, err := components.Initialize(providers.Meta, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.adAccountId = params.Metadata[metadataKeyAdAccountID]
	conn.businessId = params.Metadata[metadataKeyBusinessID]

	return conn, nil
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{Connector: base}

	// Set the metadata provider for the connector
	adapter.SchemaProvider = schema.NewObjectSchemaProvider(
		adapter.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  adapter.buildSingleObjectMetadataRequest,
			ParseResponse: adapter.parseSingleObjectMetadataResponse,
		},
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		registry,
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	return adapter, nil
}
