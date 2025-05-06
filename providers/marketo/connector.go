package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID

	*providers.ProviderInfo
	*components.URLManager
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		// The module is resolved on behalf of the user if the option is missing.
		WithModule(providers.ModuleMarketoLeads),
	)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:          params.Caller.Client,
				ResponseHandler: responseHandler,
			},
		},
		moduleID: params.Module.Selection.ID,
	}

	conn.ProviderInfo, err = providers.ReadInfo(conn.Provider(), &params.Workspace)
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

func (c *Connector) Provider() providers.Provider {
	return providers.Marketo
}

func (c *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	objectName = common.AddSuffixIfNotExists(objectName, ".json")

	return c.ModuleAPI.URL(objectName)
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
