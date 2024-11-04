package hubspot

import (
	"path"
)

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(arg string) string {
	return c.BaseURL + "/" + path.Join(c.Module.Path(), arg)
}
