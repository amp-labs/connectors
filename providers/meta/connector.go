package meta

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/meta/internal/whatsapp"
)

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	WhatsApp *whatsapp.Adapter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Init(providers.Meta, params,
		func(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
			return &Connector{Connector: base}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	switch connector.Module() { //nolint:exhaustive
	case providers.ModuleMetaWhatsApp:
		adapter, err := whatsapp.NewAdapter(params)
		if err != nil {
			return nil, err
		}

		connector.WhatsApp = adapter
	default:
		return nil, common.ErrUnsupportedModule
	}

	return connector, nil
}

func (c *Connector) Write(ctx context.Context, params connectors.WriteParams) (*connectors.WriteResult, error) {
	if c.WhatsApp != nil {
		return c.WhatsApp.Write(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
