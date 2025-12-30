package salesloft

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesloft/internal/metadata"
)

const ApiVersion = "v2"

type Connector struct {
	BaseURL string
	Module  common.Module
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.Salesloft)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
	}
	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Salesloft
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	path, _ := metadata.Schemas.LookupURLPath(c.Module.ID, objectName)
	if len(path) == 0 {
		// Fallback, try objectName as a URL.
		path = objectName
	}

	return urlbuilder.New(c.BaseURL, ApiVersion, path)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
