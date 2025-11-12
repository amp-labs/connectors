package custom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion    = "60.0"
	versionPrefix = "v"
	version       = versionPrefix + apiVersion
)

type Adapter struct {
	ClientCRM  *common.JSONHTTPClient
	XMLClient  *common.XMLHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewAdapter(
	httpClient *common.HTTPClient, salesforceCRMClient *common.JSONHTTPClient,
	moduleInfo *providers.ModuleInfo,
) *Adapter {
	return &Adapter{
		ClientCRM: salesforceCRMClient, // reuses error handling from Salesforce CRM connector.
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

func (a *Adapter) getQueryURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), "services/data", version, "query")
}

func (a *Adapter) getUserInfoURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), "services/oauth2/userinfo")
}

func (a *Adapter) getSobjectsURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), "services/data", version, "sobjects", objectName)
}
