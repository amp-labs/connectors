package salesforce

import (
	"github.com/amp-labs/connectors/catalog"
)

// Provider returns the connector provider.
func (c *Connector) Provider() catalog.Provider {
	return catalog.Salesforce
}
