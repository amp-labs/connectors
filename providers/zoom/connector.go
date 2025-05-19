package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

type Connector struct {
	BaseURL    string
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID

	*providers.ProviderInfo
	*components.URLManager
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts, WithModule(providers.ModuleZoomMeeting))
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
		moduleID: params.Module.Selection.ID,
	}

	conn.ProviderInfo, err = providers.ReadInfo(conn.Provider())
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

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.moduleID, objectName)
	if err != nil {
		return nil, err
	}

	return c.ModuleAPI.URL(path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return c.ModuleAPI.URL(path)
}

func (c *Connector) Provider() providers.Provider {
	return providers.Zoom
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
