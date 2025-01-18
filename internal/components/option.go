package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

type Option func(*ConnectorComponent)

func WithErrorHandler(handler interpreter.ErrorHandler) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON.HTTPClient.ErrorHandler = handler.Handle
		c.ClientComponent.XML.HTTPClient.ErrorHandler = handler.Handle
	}
}

func WithErrorPostProcessor(processor common.ErrorPostProcessor) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON.ErrorPostProcessor = processor
		c.ClientComponent.XML.ErrorPostProcessor = processor
	}
}

func WithResponseHandler(handler common.ResponseHandler) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON.HTTPClient.ResponseHandler = handler
		c.ClientComponent.XML.HTTPClient.ResponseHandler = handler
	}
}

func WithProviderEndpointSupport(support ProviderEndpointSupport) Option {
	return func(c *ConnectorComponent) {
		c.ProviderEndpointSupport = support
	}
}

func WithJSONHTTPClient(client *common.JSONHTTPClient) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON = client
	}
}

func WithXMLHTTPClient(client *common.XMLHTTPClient) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.XML = client
	}
}
