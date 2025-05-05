package components

import (
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
