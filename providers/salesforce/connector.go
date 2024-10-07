package salesforce

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion                 = "59.0"
	versionPrefix              = "v"
	version                    = versionPrefix + apiVersion
	restAPISuffix              = "/services/data/" + version
	uriSobjects                = restAPISuffix + "/sobjects"
	uriToolingEventRelayConfig = "tooling/sobjects/EventRelayConfig"
)

type Connector struct {
	deep.Clients
	deep.EmptyCloser
}

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.Salesforce, interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: interpretJSONError},
		XML:  &interpreter.DirectFaultyResponder{Callback: interpretXMLError},
	}).Build(opts)
}

func APIVersionSOAP() string {
	return apiVersion
}

func (c *Connector) getRestApiURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		restAPISuffix, // scope URLs to API version
	}, paths...)

	return urlbuilder.New(c.BaseURL(), parts...)
}

func (c *Connector) getDomainURL(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), paths...)
}

func (c *Connector) getSoapURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), "services/Soap/m", APIVersionSOAP())
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) getURIPartEventRelayConfig(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriToolingEventRelayConfig, paths...)
}

func (c *Connector) getURIPartSobjectsDescribe(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriSobjects, objectName, "describe")
}
