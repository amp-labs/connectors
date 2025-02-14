package front

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (*Connector, error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	jsonClient := common.JSONHTTPClient{
		HTTPClient: params.Caller,
	}

	connector := Connector{
		Client: &jsonClient,
	}

	providerInfo, err := providers.ReadInfo(connector.Provider())
	if err != nil {
		return nil, err
	}

	connector.BaseURL = providerInfo.BaseURL

	return &connector, nil
}

func (conn *Connector) Provider() providers.Provider {
	return providers.Front
}

func (conn *Connector) String() string {
	return conn.Provider() + ".Connector"
}

func (c *Connector) getAPIURL(object string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, object)
}
