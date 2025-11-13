package custom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "60.0"

type Adapter struct {
	XMLClient  *common.XMLHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewAdapter(httpClient *common.HTTPClient, moduleInfo *providers.ModuleInfo) *Adapter {
	return &Adapter{
		XMLClient: &common.XMLHTTPClient{
			HTTPClient: httpClient,
		},
		moduleInfo: moduleInfo,
	}
}

func (a *Adapter) getModuleURL() string {
	return a.moduleInfo.BaseURL
}

func (a *Adapter) getSoapURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), "services/Soap/m", apiVersion)
}
