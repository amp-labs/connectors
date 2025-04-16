package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

type Connector struct {
	BaseURL    string
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts,
		// The module is resolved on behalf of the user if the option is missing.
		WithModule(providers.ModuleZendeskTicketing),
	)
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

	providerInfo, err := providers.ReadInfo(conn.Provider(), &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo, err = providerInfo.ReadModuleInfo(conn.moduleID)
	if err != nil {
		return nil, err
	}

	// connector and its client must mirror base url and provide its own error parser
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.ZendeskSupport
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.moduleID, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	if objectsUnsupportedWrite[c.moduleID].Has(objectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if path, ok := writeURLExceptions[c.moduleID][objectName]; ok {
		// URL for write differs from read.
		return urlbuilder.New(c.BaseURL, path)
	}

	return c.getReadURL(objectName)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
