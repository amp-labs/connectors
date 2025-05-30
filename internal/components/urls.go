package components

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

// URLs manages provider-level and module-level base URLs for a connector.
//
// Embed this struct to centralize URL handling and to implement connectors.ProxyConnector automatically.
type URLs struct {
	// Base is the root API URL for the provider.
	Base string
	// Module is the base URL for the selected module (if applicable, otherwise matches Base).
	Module string
}

var _ connectors.ProxyConnector = &URLs{}

func NewURLs(
	providerInfo *providers.ProviderInfo,
	moduleInfo providers.ModuleInfo,
) *URLs {
	return &URLs{
		Base:   providerInfo.BaseURL,
		Module: moduleInfo.BaseURL,
	}
}

func (u *URLs) ProxyURL() (*url.URL, error) {
	if isTemplate(u.Base) {
		// Templates cannot be used for proxy
		return nil, common.ErrProxyNotApplicable
	}

	endpoint, err := url.Parse(u.Base)
	if err != nil {
		return nil, common.ErrProxyNotApplicable
	}

	return endpoint, nil
}

func (u *URLs) ProxyModuleURL() (*url.URL, error) {
	if isTemplate(u.Module) {
		// Templates cannot be used for proxy
		return nil, common.ErrProxyNotApplicable
	}

	endpoint, err := url.Parse(u.Module)
	if err != nil {
		return nil, common.ErrProxyNotApplicable
	}

	return endpoint, nil
}

func (u *URLs) NewBaseURL(path ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(u.Base, path...)
}

func (u *URLs) NewModuleURL(path ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(u.Module, path...)
}

// DynamicURLs is an extension of URLs for use with templated URLs containing variables like <<var>>.
//
// It supports resolving these templates dynamically per API call.
type DynamicURLs struct {
	URLs
}

var _ connectors.ProxyConnector = &DynamicURLs{}

func NewDynamicURLs(
	providerInfo *providers.ProviderInfo,
	moduleInfo providers.ModuleInfo,
) *DynamicURLs {
	return &DynamicURLs{
		URLs: *NewURLs(providerInfo, moduleInfo),
	}
}

// NewBaseURL resolves <<var>> placeholders in the base URL.
func (u *DynamicURLs) NewBaseURL(parts map[string]string, path ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(resolveTemplate(parts, u.Base), path...)
}

// NewModuleURL resolves <<var>> placeholders in the module URL.
func (u *DynamicURLs) NewModuleURL(parts map[string]string, path ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(resolveTemplate(parts, u.Module), path...)
}

// isTemplate checks if a URL contains <<var>> placeholders.
// Such URLs are not usable for proxy connections.
func isTemplate(providerURL string) bool {
	return strings.Contains(providerURL, "<<") && strings.Contains(providerURL, ">>")
}

// resolveTemplate replaces all <<var>> placeholders in the template with corresponding values.
func resolveTemplate(parts map[string]string, urlTemplate string) string {
	resolved := urlTemplate

	for key, value := range parts {
		placeholder := fmt.Sprintf("<<%v>>", key)
		resolved = strings.ReplaceAll(resolved, placeholder, value)
	}

	return resolved
}
