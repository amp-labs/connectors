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
	BaseURL    string
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
	moduleID   common.ModuleID
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parametersInternal{}, opts, WithModule(providers.ModuleKeapV1))
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
		moduleID: params.Module.Selection.ID,
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(conn.Provider())
	if err != nil {
		return nil, err
	}

	conn.moduleInfo = providerInfo.ReadModuleInfo(conn.moduleID)

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: conn.interpretHTMLError},
	}.Handle

	return conn, nil
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.moduleID, objectName)
	if err != nil {
		return nil, err
	}

	return c.getURL(path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	modulePath := metadata.Schemas.LookupModuleURLPath(c.moduleID)
	path := objectNameToWritePath.Get(objectName)

	return c.getURL(modulePath, path)
}

func (c *Connector) getURL(args ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, append([]string{
		ApiPathPrefix,
	}, args...)...)
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
