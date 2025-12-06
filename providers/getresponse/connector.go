package getresponse

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
)

type Connector struct {
	*components.Connector
	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.GetResponse, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	c := &Connector{
		Connector: base,
	}

	// static metadata for ListObjectMetadata
	c.SchemaProvider = schema.NewOpenAPISchemaProvider(
		c.ProviderContext.Module(),
		metadata.Schemas,
	)

	return c, nil
}
