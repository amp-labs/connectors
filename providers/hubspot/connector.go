package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

// Connector is a Hubspot connector.
type Connector struct {
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID

	*providers.ProviderInfo
	*components.URLManager
}

// NewConnector returns a new Hubspot connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(common.ModuleRoot), // The module is resolved on behalf of the user if the option is missing.
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
	httpClient.ErrorHandler = conn.interpretError

	// Read provider info & replace catalog variables with given substitutions, if any
	conn.ProviderInfo, err = providers.ReadInfo(providers.Hubspot)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo, err = conn.ProviderInfo.ReadModuleInfo(conn.moduleID)
	if err != nil {
		return nil, err
	}

	conn.URLManager = components.NewURLManager(conn.ProviderInfo, conn.moduleInfo)

	return conn, nil
}
