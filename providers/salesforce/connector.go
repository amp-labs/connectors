package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm"
	crmcore "github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
	"github.com/amp-labs/connectors/providers/salesforce/internal/pardot"
)

// Connector provides integration with Salesforce provider.
//
// This implementation currently supports two functional modules:
//
//   - CRM: the primary Salesforce data module responsible for standard objects.
//   - Pardot: the Account Engagement module, implemented as a separate adapter.
//
// The CRM module is undergoing partial migration: some operations are implemented directly within Connector,
// while others are delegated to specialized sub-adapters (see below).
// These sub-adapters will be consolidated as the migration completes under "crm.Adapter".
type Connector struct {
	Client *common.JSONHTTPClient

	providerInfo *providers.ProviderInfo
	moduleInfo   *providers.ModuleInfo
	moduleID     common.ModuleID

	// crmAdapter handles the core Salesforce CRM module.
	// It provides dedicated support for SalesforceCRM-specific functionality.
	crmAdapter *crm.Adapter

	// pardotAdapter handles the Salesforce Account Engagement (Pardot) module.
	// It provides dedicated support for Pardot-specific endpoints and metadata.
	pardotAdapter *pardot.Adapter
}

// NewConnector returns a new Salesforce connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(providers.ModuleSalesforceCRM),
	)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		moduleID: params.Module.Selection.ID,
	}

	conn.providerInfo, err = providers.ReadInfo(conn.Provider(), &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	// Proxy actions use the base URL set on the HTTP client, so we need to set it here.
	conn.SetBaseURL(conn.moduleInfo.BaseURL)

	// Setup CRM error handler for methods that have not been moved to internal/crm.
	conn.Client.HTTPClient.ErrorHandler = crmcore.NewErrorHandler().Handle

	// Initialize the Pardot (Account Engagement) adapter if applicable.
	// In that case, read/write/list metadata operations are delegated to it.
	moduleID := params.Module.Selection.ID
	if isPardotModule(moduleID) {
		conn.pardotAdapter, err = pardot.NewAdapter(conn.Client, conn.moduleInfo, params.Metadata.Map)
		if err != nil {
			return nil, err
		}
	} else {
		// Default Salesforce CRM module.
		connectorParams, err := newParams(opts)
		if err != nil {
			return nil, err
		}

		conn.crmAdapter, err = crm.NewAdapter(connectorParams)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Salesforce
}

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

// SetBaseURL
// TODO use components.Connector to inherit this method.
func (c *Connector) SetBaseURL(newURL string) {
	c.providerInfo.BaseURL = newURL
	c.moduleInfo.BaseURL = newURL
	c.HTTPClient().Base = newURL
}

func (c *Connector) getRestApiURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		crmcore.RestAPISuffix, // scope URLs to API version
	}, paths...)

	return urlbuilder.New(c.getModuleURL(), parts...)
}

func (c *Connector) getDomainURL(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.getModuleURL(), paths...)
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) getURLEventRelayConfig(identifier string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.getModuleURL(), crmcore.URIToolingEventRelayConfig, identifier)
}

// Gateway access to URLs.
func (c *Connector) getModuleURL() string {
	return c.moduleInfo.BaseURL
}

func (c *Connector) getURIPartSobjectsDescribe(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(crmcore.URISobjects, objectName, "describe")
}

func (c *Connector) isPardotModule() bool {
	return c.pardotAdapter != nil
}

func isPardotModule(moduleID common.ModuleID) bool {
	return moduleID == providers.ModuleSalesforceAccountEngagement ||
		moduleID == providers.ModuleSalesforceAccountEngagementDemo
}

// TODO when new approach to modules is fully done this will be obsolete.
var supportedModules = common.Modules{ // nolint:gochecknoglobals
	providers.ModuleSalesforceCRM: common.Module{
		ID: providers.ModuleSalesforceCRM,
	},
	providers.ModuleSalesforceAccountEngagement: common.Module{
		ID: providers.ModuleSalesforceAccountEngagement,
	},
	providers.ModuleSalesforceAccountEngagementDemo: common.Module{
		ID: providers.ModuleSalesforceAccountEngagementDemo,
	},
}
