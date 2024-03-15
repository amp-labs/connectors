package hubspot

import (
	"strings"
)

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(arg string) string {
	module, _ := c.ProviderInfo().GetOption(PlaceholderModule)
	baseURL := c.ProviderInfo().BaseURL

	return strings.Join([]string{baseURL, module, arg}, "/")
}
