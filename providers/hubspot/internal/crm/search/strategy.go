package search

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

type Strategy struct {
	clientCRM          *common.JSONHTTPClient
	moduleInfo         *providers.ModuleInfo
	associationsFiller associations.Filler
}

func NewStrategy(
	hubspotCRMClient *common.JSONHTTPClient,
	moduleInfo *providers.ModuleInfo,
	associationsStrategy associations.Filler,
) *Strategy {
	return &Strategy{
		clientCRM:          hubspotCRMClient, // reuses error handling from Hubspot CRM connector.
		moduleInfo:         moduleInfo,
		associationsFiller: associationsStrategy,
	}
}

func (s Strategy) getModuleURL(paths ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, paths...)
}

// https://developers.hubspot.com/docs/api-reference/search/guide
func (s Strategy) getObjectsAPISearchURL(objectName string) (*urlbuilder.URL, error) {
	return s.getModuleURL(core.APIVersion3, "objects", objectName, "search")
}

func (s Strategy) getSearchURL(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case "lists":
		// https://developers.hubspot.com/docs/api-reference/crm-lists-v3/guide#retrieve-by-searching-list-details
		return s.getModuleURL(core.APIVersion3, "lists/search")
	default:
		return nil, fmt.Errorf("%w: search not supported for %v", common.ErrObjectNotSupported, objectName)
	}
}
