package odoo

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
	components.SchemaProvider
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Odoo, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)

	return connector, nil
}

func (c *Connector) getURL(model, method string) (string, error) {
	base := c.ProviderInfo().BaseURL

	u, err := urlbuilder.New(base, "json", jsonAPIVersion, model, method)
	if err != nil {
		return "", fmt.Errorf("build odoo url: %w", err)
	}

	return u.String(), nil
}
