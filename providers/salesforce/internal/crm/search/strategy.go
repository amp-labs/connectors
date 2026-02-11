package search

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	crmcore "github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

type Strategy struct {
	clientCRM  *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewStrategy(
	salesforceCRMClient *common.JSONHTTPClient,
	moduleInfo *providers.ModuleInfo,
) *Strategy {
	return &Strategy{
		clientCRM:  salesforceCRMClient, // reuses error handling from Salesforce CRM connector.
		moduleInfo: moduleInfo,
	}
}

func (s Strategy) getModuleURL(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, paths...)
}

// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_query.htm
func (s Strategy) getQueryURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, crmcore.RestAPISuffix, "query")
}
