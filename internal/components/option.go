package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

type Option func(*ConnectorComponent)

func WithErrorHandler(handler interpreter.ErrorHandler) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON.HTTPClient.ErrorHandler = handler.Handle
		c.ClientComponent.xml.HTTPClient.ErrorHandler = handler.Handle
	}
}

func WithErrorPostProcessor(processor common.ErrorPostProcessor) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON.ErrorPostProcessor = processor
		c.ClientComponent.xml.ErrorPostProcessor = processor
	}
}

func WithResponseHandler(handler common.ResponseHandler) Option {
	return func(c *ConnectorComponent) {
		c.ClientComponent.JSON.HTTPClient.ResponseHandler = handler
		c.ClientComponent.xml.HTTPClient.ResponseHandler = handler
	}
}

func WithMetadataStrategy(strategy MetadataStrategy) Option {
	return func(c *ConnectorComponent) {
		c.MetadataStrategy = strategy
	}
}

func WithReadStrategy(strategy ReadStrategy) Option {
	return func(c *ConnectorComponent) {
		c.ReadStrategy = strategy
	}
}

func WithWriteStrategy(strategy WriteStrategy) Option {
	return func(c *ConnectorComponent) {
		c.WriteStrategy = strategy
	}
}

func WithProviderEndpointSupport(support ProviderEndpointSupport) Option {
	return func(c *ConnectorComponent) {
		c.ProviderEndpointSupport = support
	}
}
