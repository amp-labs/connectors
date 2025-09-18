package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

const apiVersion = "/api/v2"

// Connector covers ticketing as well as help center.
// https://developer.zendesk.com/api-reference/ticketing/introduction/
// https://developer.zendesk.com/api-reference/help_center/help-center-api/introduction/
type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	httpClient := params.Client.Caller
	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: httpClient,
		},
	}

	providerInfo, err := providers.ReadInfo(conn.Provider(), &params.Workspace)
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
	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.BaseURL, apiVersion, path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	if objectsUnsupportedWrite[common.ModuleRoot].Has(objectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if path, ok := writeURLExceptions[common.ModuleRoot][objectName]; ok {
		// URL for write differs from read.
		return urlbuilder.New(c.BaseURL, apiVersion, path)
	}

	return c.getReadURL(objectName)
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
