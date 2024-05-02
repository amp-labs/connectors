package gong

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"

	"errors"

	"github.com/amp-labs/connectors/providers"
)

var DefaultModule = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "api/data",
	Version: "v2",
}

type Connector struct {
	BaseURL   string
	Client    *common.JSONHTTPClient
	APIModule paramsbuilder.APIModule
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
		Client:    params.client,
		BaseURL:   providerInfo.BaseURL,
		APIModule: params.APIModule,
	}

	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: func(res *http.Response, body []byte) error {
			// You need to create an error from the response and body
			// This is just an example, adjust it according to your needs
			err := errors.ErrUnsupported
			return conn.HandleError(err)
		},
	}.Handle

	return conn, nil

}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
