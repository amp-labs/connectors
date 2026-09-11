package ramp

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/ramp/metadata"
)

type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader
	components.Writer
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Ramp, params, constructor)
}

func NewDemoConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.RampDemo, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	return connector, nil
}
