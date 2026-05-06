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

// Connector provides integration with Hubspot provider.
//
// The CRM module is undergoing partial migration: some operations are implemented directly within Connector,
// while others are delegated to specialized sub-adapters (see below).
// These sub-adapters will be consolidated as the migration completes under "crm.Adapter".
type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// crmAdapter handles the core Hubspot CRM module.
	// It provides dedicated support for HubspotCRM-specific functionality.
	crmAdapter *crm.Adapter
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
	if connector.Module() == providers.ModuleHubspotCRM {
		connector.crmAdapter, err = crm.NewAdapter(params)
		if err != nil {
			return nil, err
		}
	}

	return connector, nil
}

func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.Search(ctx, params)
	}

	return nil, common.ErrNotImplemented
}
