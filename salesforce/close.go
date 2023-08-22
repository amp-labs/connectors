package salesforce

func (c *Connector) Close() error {
	c.client.CloseIdleConnections()

	return nil
}
