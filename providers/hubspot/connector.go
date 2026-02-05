package hubspot

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/batch"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm"
	"github.com/amp-labs/connectors/providers/hubspot/internal/custom"
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

	// CRM module sub-adapters
	// These delegate specialized subsets of Hubspot CRM functionality to keep Connector modular and prevent code bloat.
	customAdapter *custom.Adapter // used for connectors.UpsertMetadataConnector capabilities.
	batchAdapter  *batch.Adapter  // used for connectors.BatchWriteConnector capabilities.
	crmAdapter    *crm.Adapter    // used for connectors.DeleteConnector capabilities.
}

const (
	ModuleCRMVersion = "v3"
)

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

	conn.customAdapter = custom.NewAdapter(conn.Client, conn.moduleInfo)
	conn.batchAdapter = batch.NewAdapter(conn.Client.HTTPClient, conn.moduleInfo)

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
