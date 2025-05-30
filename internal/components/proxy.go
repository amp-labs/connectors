package components

import (
	"net/url"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// DefaultProxy is a transitional implementation of the connectors.ProxyConnector interface.
//
// It wraps an existing Connector instance and exposes the provider-level proxy URL.
// This reflects the current de facto behavior in connectors that reach into HTTPClient().Base directly.
//
// Usage of DefaultProxy allows step-by-step migration:
//  1. Enforce the ProxyConnector interface.
//  2. Replace connector-specific logic with DefaultProxy.
//  3. Eventually replace DefaultProxy with the proper URLs-based implementation,
//     which will support both provider and module-level proxying.
type DefaultProxy struct {
	connector connectors.Connector
}

var _ connectors.ProxyConnector = &DefaultProxy{}

// NewDefaultProxy returns a DefaultProxy wrapper around the given connector.
//
// This should be embedded into the connector struct to expose connectors.ProxyConnector behavior.
func NewDefaultProxy(connector connectors.Connector) *DefaultProxy {
	return &DefaultProxy{connector: connector}
}

func (c *DefaultProxy) ProxyURL() (*url.URL, error) {
	if c.connector == nil {
		return nil, common.ErrNotImplemented
	}

	return url.Parse(c.connector.HTTPClient().Base)
}

// ProxyModuleURL is not supported in DefaultProxy and always returns ErrProxyNotApplicable.
func (c *DefaultProxy) ProxyModuleURL() (*url.URL, error) {
	if c.connector == nil {
		return nil, common.ErrNotImplemented
	}

	return nil, common.ErrProxyNotApplicable
}
