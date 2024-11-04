package hubspot

import (
	"path"
)

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(arg string) string {
	// TODO: use url package to join paths and avoid issues with slashes in another PR
	return c.BaseURL + "/" + path.Join(c.Module.Path(), arg)
}
