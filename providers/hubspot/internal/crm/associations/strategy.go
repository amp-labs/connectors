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

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/batch/post-crm-v4-associations-fromObjectType-toObjectType-batch-read
func (s Strategy) getAssociationsURL(fromObject, toObject string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, core.APIVersion, "associations", fromObject, toObject, "batch/read")
}
