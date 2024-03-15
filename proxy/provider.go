package proxy

import (
	"github.com/amp-labs/connectors/providers"
)

func (c *Connector) Provider() providers.Provider {
	return c.provider
}

func (c *Connector) ProviderInfo() *providers.ProviderInfo {
	return c.providerInfo
}
