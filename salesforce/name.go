package salesforce

// Name returns the name of the connector.
func (c *Connector) Name() string {
	return c.Provider().String()
}
