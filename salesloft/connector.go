package salesloft

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

var DefaultModuleCRM = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Version: "v2",
}

type Connector struct {
	BaseURL string
	Module  string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := parameters{}.FromOptions(opts...)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.Salesloft, nil)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Module: params.Module.Suffix,
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
	}
	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: conn.interpretJSONError,
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.Salesloft
}

func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector[%s]", c.Provider(), c.Module)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	parts := []string{c.BaseURL, c.Module, arg}
	filtered := make([]string, 0)

	for _, part := range parts {
		if len(part) != 0 {
			filtered = append(filtered, part)
		}
	}

	return urlbuilder.New(strings.Join(filtered, "/"))
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
