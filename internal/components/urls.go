package components

import (
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// URLs manages provider-level and module-level base URLs for a connector.
//
// Embed this struct to centralize URL handling and to implement connectors.ProxyConnector automatically.
type URLs struct {
	providerContext ProviderContext
}

var _ connectors.ProxyConnector = &URLs{}

// NewURLs constructs an entity that holds proxy URLs.
func NewURLs(connector *Connector) *URLs {
	return &URLs{
		providerContext: connector.ProviderContext,
	}
}

// ProxyURL returns URL that can be used for general purpose.
func (u *URLs) ProxyURL() (*url.URL, error) {
	proxyURL, err := url.Parse(u.providerContext.providerInfo.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", common.ErrProxyNotApplicable, err)
	}

	return proxyURL, nil
}

// ProxyModuleURL returns URL which is module specific.
// When provider doesn't have any modules this acts identical to ProxyURL.
func (u *URLs) ProxyModuleURL() (*url.URL, error) {
	proxyURL, err := url.Parse(u.providerContext.moduleInfo.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", common.ErrProxyNotApplicable, err)
	}

	return proxyURL, nil
}
