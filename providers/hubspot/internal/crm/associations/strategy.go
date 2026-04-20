package associations

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

type Strategy struct {
	clientCRM    *common.JSONHTTPClient
	moduleInfo   *providers.ModuleInfo
	providerInfo *providers.ProviderInfo
}

func NewStrategy(
	hubspotCRMClient *common.JSONHTTPClient, moduleInfo *providers.ModuleInfo, providerInfo *providers.ProviderInfo,
) *Strategy {
	return &Strategy{
		clientCRM:    hubspotCRMClient,
		moduleInfo:   moduleInfo,
		providerInfo: providerInfo,
	}
}

// getReadAssociationsURL builds the Hubspot endpoint to batch read associations between 2 object types.
//
// nolint:lll
// https://developers.hubspot.com/docs/api-reference/latest/crm/associations/associate-records/batch/get-associations
func (s Strategy) getReadAssociationsURL(fromObject, toObject string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL,
		"associations", core.APIVersion2026March, fromObject, toObject, "batch/read")
}

// getCreateAssociationsURL builds the Hubspot endpoint to create associations between 2 object types.
//
// nolint:lll
// https://developers.hubspot.com/docs/api-reference/latest/crm/associations/associate-records/batch/create-associations-labeled
func (s Strategy) getCreateAssociationsURL(fromObject, toObject string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL,
		"associations", core.APIVersion2026March, fromObject, toObject, "batch/create")
}

// getReadObjectSchema builds the Hubspot endpoint to get schema definition for an object.
// Response includes properties and associations defined on that object.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/schemas/get-schema
func (s Strategy) getReadObjectSchema(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, "crm-object-schemas", core.APIVersion2026March, "schemas", objectName)
}
