package facebook

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Adapter struct {
	// Basic connector
	*components.Connector

	common.RequireMetadata
	// Supported operations
	components.SchemaProvider

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

	return adapter, nil
}
