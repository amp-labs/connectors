package custom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
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
// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/batch/post-crm-v3-properties-objectType-batch-create
func (a *Adapter) getPropertyBatchCreateURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, "/batch/create")
}

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/core/patch-crm-v3-properties-objectType-propertyName
func (a *Adapter) getPropertyUpdateURL(objectName, propertyName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, propertyName)
}

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/groups/get-crm-v3-properties-objectType-groups-groupName
func (a *Adapter) getPropertyGroupNameURL(objectName, groupName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, "groups", groupName)
}

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/groups/post-crm-v3-properties-objectType-groups
func (a *Adapter) getPropertyGroupNameCreationURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, "groups")
}
