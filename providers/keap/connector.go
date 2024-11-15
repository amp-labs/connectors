package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

const ApiPathPrefix = "crm/rest"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
	Module  common.Module
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := paramsbuilder.Apply(parameters{}, opts, WithModule(ModuleV1))
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		Module: params.Selection,
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: conn.interpretHTMLError},
	}.Handle

	return conn, nil
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module.ID, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, ApiPathPrefix, path)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}

func (c *Connector) Provider() providers.Provider {
	return providers.Keap
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
