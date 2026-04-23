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

// https://developers.hubspot.com/docs/api-reference/latest/crm/properties/batch/create-properties
// Note: Version APIVersion2026March is NOT FOUND at the moment for this endpoint. Using older V3.
func (a *Adapter) getPropertyBatchCreateURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, "/batch/create")
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/properties/update-property
// Note: Version APIVersion2026March is NOT FOUND at the moment for this endpoint. Using older V3.
func (a *Adapter) getPropertyUpdateURL(objectName, propertyName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, propertyName)
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/properties/property-groups/get-property
// Note: Version APIVersion2026March is NOT FOUND at the moment for this endpoint. Using older V3.
func (a *Adapter) getPropertyGroupNameURL(objectName, groupName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, "groups", groupName)
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/properties/property-groups/create-property
// Note: Version APIVersion2026March is NOT FOUND at the moment for this endpoint. Using older V3.
func (a *Adapter) getPropertyGroupNameCreationURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.moduleInfo.BaseURL, core.APIVersion3, "properties", objectName, "groups")
}
