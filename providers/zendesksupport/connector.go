package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

type Connector struct {
	BaseURL    string
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID

	*providers.ProviderInfo
	*components.URLManager
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
	httpClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle

	conn.ProviderInfo, err = providers.ReadInfo(conn.Provider(), &params.Workspace)
	if err != nil {
		return nil, err
	}

	conn.moduleInfo, err = conn.ProviderInfo.ReadModuleInfo(conn.moduleID)
	if err != nil {
		return nil, err
	}

	conn.URLManager = components.NewURLManager(conn.ProviderInfo, conn.moduleInfo)

	return conn, nil
}

func (c *Connector) Provider() providers.Provider {
	return providers.ZendeskSupport
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.moduleID, objectName)
	if err != nil {
		return nil, err
	}

	return c.ModuleAPI.URL(path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	if objectsUnsupportedWrite[c.moduleID].Has(objectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if path, ok := writeURLExceptions[c.moduleID][objectName]; ok {
		// URL for write differs from read.
		return c.ModuleAPI.URL(path)
	}

	return c.getReadURL(objectName)
}
