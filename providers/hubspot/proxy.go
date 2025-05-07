package hubspot

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) ProxyURL() (*url.URL, error) {
	return url.Parse(c.HTTPClient().Base)
}

func (c *Connector) ProxyModuleURL() (*url.URL, error) {
	return nil, common.ErrProxyNotApplicable
}
