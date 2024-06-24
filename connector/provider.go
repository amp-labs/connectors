package connector

import (
	"github.com/amp-labs/connectors/catalog"
)

func (c *Connector) Provider() catalog.Provider {
	return c.provider
}
