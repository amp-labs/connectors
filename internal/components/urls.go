package components

import (
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// ProxyResolver derives proxy configurations from provider and module metadata.
type ProxyResolver struct {
	defaultURL string
	moduleURL  string
}

// NewProxyResolver constructs a ProxyResolver using the given provider context.
func NewProxyResolver(providerContext ProviderContext) *ProxyResolver {
	return &ProxyResolver{
		defaultURL: providerContext.ProviderInfo().BaseURL,
		moduleURL:  providerContext.ModuleInfo().BaseURL,
	}
}

// ProxyConfig returns the general proxy configuration for the provider.
//
// It uses the provider-level BaseURL and does not apply any module-specific behavior.
//
// Returns common.ErrProxyNotApplicable if the URL cannot be parsed.
func (r *ProxyResolver) ProxyConfig() (*connectors.ProxyConfig, error) {
	_, err := url.Parse(r.defaultURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", common.ErrProxyNotApplicable, err)
	}

	return &connectors.ProxyConfig{
		URL: r.defaultURL,
	}, nil
}

// ProxyModuleConfig returns the module-specific proxy configuration.
//
// It uses the module-level BaseURL. If the instance is not associated
// with a module, this behaves the same as ProxyConfig().
//
// Returns common.ErrProxyNotApplicable if the URL cannot be parsed.
func (r *ProxyResolver) ProxyModuleConfig() (*connectors.ProxyConfig, error) {
	_, err := url.Parse(r.moduleURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", common.ErrProxyNotApplicable, err)
	}

	return &connectors.ProxyConfig{
		URL: r.moduleURL,
	}, nil
}
