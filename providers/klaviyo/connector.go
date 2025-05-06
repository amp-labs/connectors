package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
)

type Connector struct {
	Client     *common.JSONHTTPClient
	moduleInfo providers.ModuleInfo
	moduleID   common.ModuleID

	*providers.ProviderInfo
	*components.URLManager
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts, WithModule(providers.ModuleKlaviyo2024Oct15))
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
		Custom: map[string]interpreter.FaultyResponseHandler{
			"application/vnd.api+json": interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
		},
	}.Handle

	conn.ProviderInfo, err = providers.ReadInfo(conn.Provider())
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

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.moduleID, objectName)
	if err != nil {
		return nil, err
	}

	return c.RootAPI.URL("api", path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return c.RootAPI.URL(path)
}

func (c *Connector) getDeleteURL(objectName string) (*urlbuilder.URL, error) {
	return c.RootAPI.URL("api", objectName)
}

func (c *Connector) revisionHeader() common.Header {
	return common.Header{
		Key:   "revision",
		Value: string(c.moduleID),
	}
}

func (c *Connector) Provider() providers.Provider {
	return providers.Klaviyo
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
