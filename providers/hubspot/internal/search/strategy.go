package search

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

type Strategy struct {
	clientCRM          *common.JSONHTTPClient
	providerInfo       *providers.ProviderInfo
	associationsFiller associations.Filler
}

func NewStrategy(
	hubspotCRMClient *common.JSONHTTPClient,
	providerInfo *providers.ProviderInfo,
	associationsStrategy associations.Filler,
) *Strategy {
	return &Strategy{
		clientCRM:          hubspotCRMClient, // reuses error handling from Hubspot CRM connector.
		providerInfo:       providerInfo,
		associationsFiller: associationsStrategy,
	}
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/search-the-crm#make-a-search-request
func (s Strategy) getObjectsAPISearchURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, "crm", "objects", core.APIVersion2026March, objectName, "search")
}

func (s Strategy) getSearchURL(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case "lists":
		// https://developers.hubspot.com/docs/api-reference/latest/crm/lists/guide#retrieve-by-searching-list-details
		return urlbuilder.New(s.providerInfo.BaseURL, "crm", "lists", core.APIVersion2026March, "search")
	default:
		return nil, fmt.Errorf("%w: search not supported for %v", common.ErrObjectNotSupported, objectName)
	}
}
