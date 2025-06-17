package dynamicsbusiness

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

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireWorkspace
	common.RequireMetadata

	components.SchemaProvider

	environmentName string
	tenantID        string
	companyID       string
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
	}

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
