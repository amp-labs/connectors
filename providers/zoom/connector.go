package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

const (
	apiVersion     = "/v2"
	ZoomDateFormat = "2006-01-02"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Zoom
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, apiVersion, path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return urlbuilder.New(c.BaseURL, apiVersion, path)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
