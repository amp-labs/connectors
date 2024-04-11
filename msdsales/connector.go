package msdsales

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/reqrepeater"
	"github.com/amp-labs/connectors/providers"
)

var DefaultRequestRetryLimit = 3

var DefaultModuleCRM = paramsbuilder.APIModule{ // nolint: gochecknoglobals
	Label:   "api/data",
	Version: "v9.2",
}

type Connector struct {
	BaseURL       string
	Module        string
	Client        *common.JSONHTTPClient
	RetryStrategy reqrepeater.Strategy
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

	providerInfo, err := providers.ReadInfo(providers.MicrosoftDynamics365Sales, &map[string]string{
		"workspace": params.Workspace.Name,
	})
	if err != nil {
		return nil, err
	}

	baseURL := providerInfo.BaseURL
	conn = &Connector{
		BaseURL: baseURL,
		Module:  params.Module.Suffix,
		Client:  params.Client.Caller,
		RetryStrategy: &reqrepeater.UniformRetryStrategy{ // FIXME call retry strategy could be part of options
			RetryLimit: DefaultRequestRetryLimit,
			Interval:   time.Second,
		},
	}
	// connector and its client must mirror base url and provide its own error parser
	conn.Client.HTTPClient.Base = baseURL
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: conn.interpretJSONError,
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.MicrosoftDynamics365Sales
}

func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector[%s]", c.Provider(), c.Module)
}

func (c *Connector) getURL(arg string) string { // FIXME should it be part of Connector interface?
	parts := []string{c.BaseURL, c.Module, arg}
	filtered := make([]string, 0)

	for _, part := range parts {
		if len(part) != 0 {
			filtered = append(filtered, part)
		}
	}

	return strings.Join(filtered, "/")
}
