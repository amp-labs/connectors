package custom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	ModuleCRMVersion = "v3"
)

type Adapter struct {
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewAdapter(client *common.JSONHTTPClient, moduleInfo *providers.ModuleInfo) *Adapter {
	return &Adapter{
		Client:     client,
		moduleInfo: moduleInfo,
	}
}

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/properties/post-crm-properties-v3-objectType-batch-create
func (a *Adapter) getPropertyBatchCreateURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, "properties", ModuleCRMVersion, objectName, "/batch/create")
}

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/properties/patch-crm-properties-v3-objectType-propertyName
func (a *Adapter) getPropertyUpdateURL(objectName, propertyName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, "properties", ModuleCRMVersion, objectName, propertyName)
}
