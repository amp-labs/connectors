package servicenow

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	BaseURL string
	Module  common.Module
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (*Connector, error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	jsonClient := common.JSONHTTPClient{HTTPClient: httpClient}

	conn := &Connector{
		Client: &jsonClient,
		Module: params.Module.Selection,
	}

	providerInfo, err := providers.ReadInfo(conn.Provider(), &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn.setBaseURL(providerInfo.BaseURL)

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.ServiceNow
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	// https://{{.workspace}}.servicenow.com/api/now/v2/table/{objectName}
	return urlbuilder.New(c.BaseURL, restAPIPrefix, c.Module.Path(), objectName)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
