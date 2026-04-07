package associations

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

type Strategy struct {
	clientCRM  *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewStrategy(hubspotCRMClient *common.JSONHTTPClient, moduleInfo *providers.ModuleInfo) *Strategy {
	return &Strategy{
		clientCRM:  hubspotCRMClient,
		moduleInfo: moduleInfo,
	}
}

// getReadAssociationsURL builds the Hubspot endpoint to batch read associations between 2 object types.
//
// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/batch/post-crm-v4-associations-fromObjectType-toObjectType-batch-read
func (s Strategy) getReadAssociationsURL(fromObject, toObject string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, core.APIVersion, "associations", fromObject, toObject, "batch/read")
}

// getCreateAssociationsURL builds the Hubspot endpoint to create associations between 2 object types.
//
// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/batch/post-crm-v4-associations-fromObjectType-toObjectType-batch-create
func (s Strategy) getCreateAssociationsURL(fromObject, toObject string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, core.APIVersion, "associations", fromObject, toObject, "batch/create")
}

// getReadObjectSchema builds the Hubspot endpoint to get schema definition for an object.
// Response includes properties and associations defined on that object.
//
// https://developers.hubspot.com/docs/api-reference/crm-schemas-v3/core/get-crm-object-schemas-v3-schemas-objectType
func (s Strategy) getReadObjectSchema(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, core.APIVersion3, "schemas", objectName)
}
