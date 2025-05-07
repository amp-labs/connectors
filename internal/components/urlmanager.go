package components

import (
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type URLManager struct {
	RootAPI   *urlbuilder.Template
	ModuleAPI *urlbuilder.Template
}

func NewURLManager(providerInfo *providers.ProviderInfo, moduleInfo providers.ModuleInfo) *URLManager {
	providerBaseURL := providerInfo.BaseURL
	moduleBaseURL := moduleInfo.BaseURL

	return &URLManager{
		RootAPI:   urlbuilder.NewTemplate(providerBaseURL),
		ModuleAPI: urlbuilder.NewTemplate(moduleBaseURL),
	}
}

func (m URLManager) ProxyURL() (*url.URL, error) {
	endpoint, err := m.RootAPI.URL()
	if err != nil {
		// The root URL for the provider cannot have dynamic, template based URL.
		return nil, fmt.Errorf("%w: provider without static base url", common.ErrProxyNotApplicable)
	}

	return endpoint.ToURL()
}

func (m URLManager) ProxyModuleURL() (*url.URL, error) {
	endpoint, err := m.ModuleAPI.URL()
	if err != nil {
		return nil, fmt.Errorf("%w: module without static base url", common.ErrProxyNotApplicable)
	}

	return endpoint.ToURL()
}
