package salesforce

import "fmt"

// String returns a string representation of the connector, which is useful for logging / debugging.
func (c *Connector) String() string {
	return fmt.Sprintf("salesforce.Connector[%s]", c.Domain)
}
