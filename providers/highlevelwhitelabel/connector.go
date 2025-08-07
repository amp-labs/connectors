package highlevelwhitelabel

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
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
	common.RequireMetadata

	// Supported operations
	components.SchemaProvider

	locationId string
}

const (
	metadataKeyLocationID = "locationId"
)

// In the highlevel connector, there is two listing type -- standard and whitelabel.
// Both type have different Authorization URL
// refer: https://highlevel.stoplight.io/docs/integrations/a04191c0fabf9-authorization#3-get-the-apps-authorization-page-url.
// Apps are visible at both agency and sub-account levels under white-label and non-white-label domains
// (no HighLevel/GoHighLevel references allowed, or they’ll be disapproved) for white-label.
//
//	While standard (non-white-label) apps are visible under the HighLevel domain based on
//	distribution type but not under white-label domains.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// Create base connector with provider info
	conn, err := components.Initialize(providers.HighLevelWhiteLabel, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.locationId = params.Metadata[metadataKeyLocationID]

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
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

	return connector, nil
}
