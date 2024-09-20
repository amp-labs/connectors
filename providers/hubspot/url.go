package hubspot

import (
	"strings"
)

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(arg string) string {
	return strings.Join([]string{c.BaseURL, c.Module.Path(), arg}, "/")
}
