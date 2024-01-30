package hubspot

import (
	"strings"
)

// getUrl is a helper to return the full URL considering the base URL & module.
func (c *Connector) getUrl(arg string) string {
	return strings.Join([]string{c.BaseURL, c.Module, arg}, "/")
}
