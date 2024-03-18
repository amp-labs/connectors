package basic

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}
