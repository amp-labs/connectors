package gong

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
