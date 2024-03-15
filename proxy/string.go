package proxy

import (
	"fmt"
)

// TODO: Extend this to allow custom labels from providerInfo for additional information like module, domain, etc.
func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector", c.Provider())
}
