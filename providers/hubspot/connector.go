package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

// Connector is a Hubspot connector.
type Connector struct {
	Client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo
	moduleInfo   *providers.ModuleInfo
	moduleID     common.ModuleID
}

// NewConnector returns a new Hubspot connector.
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

	// Read provider info & replace catalog variables with given substitutions, if any
	conn.providerInfo, err = providers.ReadInfo(providers.Hubspot)
	if err != nil {
		return nil, err
	}

	// HTTPClient will soon not store Base URL.
	conn.Client.HTTPClient.Base = conn.providerInfo.BaseURL
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError

	conn.moduleInfo = conn.providerInfo.ReadModuleInfo(conn.moduleID)

	return conn, nil
}

// This method must be used only by the unit tests.
func (c *Connector) setBaseURL(newURL string) {
	c.providerInfo.BaseURL = newURL
	c.moduleInfo.BaseURL = newURL
}
