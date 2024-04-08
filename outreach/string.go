package outreach

import "fmt"

func (c *Connector) String() string {
	return fmt.Sprintf("%s.Connector", c.Provider())
}
