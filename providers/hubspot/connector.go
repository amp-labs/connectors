package hubspot

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

// Connector implements the HubSpot integration.
//
// HubSpot is modeled as a single connector.
// Unlike multi-module providers, all supported operations belong directly to Connector.
//
// Complex feature sets may still live in dedicated internal packages
// (for example batch, search, or subscriptions), but those packages are
// implementation details rather than top-level modules.
//
// The long-term direction is for Connector to own all operations directly,
// with reusable feature strategies embedded as needed.
type Connector struct {
	// Shared connector infrastructure.
	*components.Connector

	// Provides access to an authenticated client.
	common.RequireAuthenticatedClient

	// Temporary grouping.
	// TODO this adapter contents should dissolve into this connector.
	// The idea of CRM module should cease to exist. Hubspot is but one entity.
	delegate *crm.Adapter
}

var _ connectors.WebhookVerifierConnector = &Connector{}

// NewConnector returns a new Hubspot connector.
// Hubspot connector still owns CRM functionality. Not every CRM feature is located under `crm` package.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Hubspot, params, func(base *components.Connector) (*Connector, error) {
		return constructor(base, &params)
	})
}

func constructor(base *components.Connector, params *common.ConnectorParams) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	// Note: error handler must return common.HTTPError.
	// Check method in the internal package "custom", method "readGroupName" which relies on error casting.
	connector.SetErrorHandler(core.InterpretJSONError)

	var err error

	connector.delegate, err = crm.NewAdapter(params)
	if err != nil {
		return nil, err
	}

	return connector, nil
}

func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	return c.delegate.Search(ctx, params)
}
