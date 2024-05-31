package gong

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const ApiVersion = "v2"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func WithCatalogSubstitutions(substitutions map[string]string) Option {
	return func(params *gongParams) {
		params.substitutions = substitutions
	}
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params := &gongParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(providers.Gong, &map[string]string{
		"workspace": params.Workspace.Name,
	})
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client:  params.client,
		BaseURL: providerInfo.BaseURL,
	}

	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError

	return conn, nil
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) interpretError(res *http.Response, body []byte) error {
	return common.InterpretError(res, body)
}
