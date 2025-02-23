package helpscout

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const restAPIVersion = "v2"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (*Connector, error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.HelpScoutMailbox)
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
	return providers.HelpScoutMailbox
}

func (conn *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(conn.BaseURL, restAPIVersion, objectName)
}

func (conn *Connector) setBaseURL(newURL string) {
	conn.BaseURL = newURL
	conn.Client.HTTPClient.Base = newURL
}
