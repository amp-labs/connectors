package hubspot

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/custom"
)

// Connector is a Hubspot connector.
type Connector struct {
	Client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo
	moduleInfo   *providers.ModuleInfo
	moduleID     common.ModuleID

	// Delegate for the UpsertMetadat functionality.
	customAdapter *custom.Adapter
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
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError
	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	conn.customAdapter = custom.NewAdapter(conn.Client, conn.moduleInfo)

	return conn, nil
}
