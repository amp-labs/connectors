package g2

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider

	// productId represents either subject_product_id or product_id.
	productId string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.G2, params, constructor)
	if err != nil {
		return nil, err
	}

	connector.productId = params.Metadata["productId"]

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"productId"},
		},
	}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
		},
	)

	return connector, nil
}
