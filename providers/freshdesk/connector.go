package freshdesk

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const restAPIPrefix = "api/v2"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (*Connector, error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.Freshdesk, &params.Workspace)
	if err != nil {
		return nil, err
	}

	jsonClient := common.JSONHTTPClient{
		HTTPClient: params.Caller,
	}

	connector := Connector{
		Client: &jsonClient,
	}

	connector.setBaseURL(providerInfo.BaseURL)

	return &connector, nil
}

func (conn *Connector) Provider() providers.Provider {
	return providers.Freshdesk
}

func (conn *Connector) String() string {
	return conn.Provider() + ".Connector"
}

func (conn *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	path, exists := objectResourcePath[objectName]
	if !exists {
		path = objectName
	}

	return urlbuilder.New(conn.BaseURL, restAPIPrefix, path)
}

func (conn *Connector) setBaseURL(newURL string) {
	conn.BaseURL = newURL
	conn.Client.HTTPClient.Base = newURL
}
