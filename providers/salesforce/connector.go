package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion                 = "59.0"
	versionPrefix              = "v"
	version                    = versionPrefix + apiVersion
	restAPISuffix              = "/services/data/" + version
	uriSobjects                = restAPISuffix + "/sobjects"
	uriToolingEventRelayConfig = restAPISuffix + "/tooling/sobjects/EventRelayConfig"
)

// Connector is a Salesforce connector.
type Connector struct {
	// Basic connector
	*components.Connector

	XMLClient *common.XMLHTTPClient
}

// NewConnector is an old constructor, use NewConnectorV2.
// Deprecated.
func NewConnector(opts ...Option) (*Connector, error) {
	params, err := newParams(opts)
	if err != nil {
		return nil, err
	}

	return NewConnectorV2(*params)
}

func NewConnectorV2(params common.Parameters) (*Connector, error) {
	return components.Initialize(providers.Salesforce, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	conn := &Connector{Connector: base}

	conn.SetErrorHandler(interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: conn.interpretJSONError},
		XML:  &interpreter.DirectFaultyResponder{Callback: conn.interpretXMLError},
	}.Handle)

	conn.XMLClient = &common.XMLHTTPClient{
		HTTPClient: base.HTTPClient(),
	}

	return conn, nil
}

func APIVersionSOAP() string {
	return apiVersion
}

func (c *Connector) getRestApiURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{
		restAPISuffix, // scope URLs to API version
	}, paths...)

	return c.RootClient.URL(parts...)
}

func (c *Connector) getDomainURL(paths ...string) (*urlbuilder.URL, error) {
	return c.RootClient.URL(paths...)
}

func (c *Connector) getSoapURL() (*urlbuilder.URL, error) {
	return c.RootClient.URL("services/Soap/m", APIVersionSOAP())
}

// nolint: lll
// https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_eventrelayconfig.htm?q=EventRelayConfig
func (c *Connector) getURIPartEventRelayConfig(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriToolingEventRelayConfig, paths...)
}

func (c *Connector) getURIPartSobjectsDescribe(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(uriSobjects, objectName, "describe")
}
