package connector

import (
	"github.com/amp-labs/connectors/providers"
)

func (c *Connector) Provider() providers.Provider {
	return c.provider
}
