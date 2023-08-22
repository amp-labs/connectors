package salesforce

const Name = "salesforce"

// Name returns the name of the connector.
func (c *Connector) Name() string {
	return Name
}
