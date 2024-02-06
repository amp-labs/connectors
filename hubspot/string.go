package hubspot

import "fmt"

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector[%s]", c.Provider(), c.Module)
}

// Name returns the name of the connector.
func (c *Connector) Name() string {
	return c.Provider().String()
}
