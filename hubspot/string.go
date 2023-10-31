package hubspot

import "fmt"

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Connector) String() string {
	return fmt.Sprintf("hubspot.Connector[%s]", c.Module)
}
