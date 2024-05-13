package docusign

import "github.com/amp-labs/connectors/providers"

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Docusign
}
