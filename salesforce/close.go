package salesforce

func (c *Connector) Close() error {
	c.Client.CloseIdleConnections()

	return nil
}
