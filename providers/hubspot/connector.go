package hubspot

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
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
	Client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo
	moduleInfo   *providers.ModuleInfo
	moduleID     common.ModuleID

	// crmAdapter handles the core Hubspot CRM module.
	// It provides dedicated support for HubspotCRM-specific functionality.
	crmAdapter *crm.Adapter
}

var _ connectors.WebhookVerifierConnector = &Connector{}

// NewConnector returns a new Hubspot connector.
// Nearly all of the logic for this connector assumes that the module is CRM (url construction, etc)
// When we have to add support for other modules, it might be best to create a separate internal package.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(common.ModuleRoot), // The module is resolved on behalf of the user if the option is missing.
	)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
		moduleID: params.Module.Selection.ID,
	}

	conn.providerInfo, err = providers.ReadInfo(providers.Hubspot)
	if err != nil {
		return nil, err
	}

	conn.Client.HTTPClient.Base = conn.providerInfo.BaseURL
	// Note: error handler must return common.HTTPError.
	// Check method in the internal package "custom", method "readGroupName" which relies on error casting.
	conn.Client.HTTPClient.ErrorHandler = core.InterpretJSONError
	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	connectorParams, err := newParams(opts)
	if err != nil {
		return nil, err
	}

	conn.crmAdapter, err = crm.NewAdapter(connectorParams)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	// Delegated.
	return c.crmAdapter.Search(ctx, params)
}
