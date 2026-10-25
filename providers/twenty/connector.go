package twenty

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

// Connector reaches one Twenty instance, self-hosted or Cloud, with a workspace API key.
// The catalog's `workspace` input is the instance's domain, which every request is sent to.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireWorkspace

	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Twenty, params, constructor)
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Used by the metadata requests made through JSONHTTPClient.
	connector.HTTPClient().ErrorHandler = errorHandler

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)

	return connector, nil
}
