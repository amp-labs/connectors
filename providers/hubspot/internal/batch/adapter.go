package batch

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

// Adapter handles batched record operations (create/update) against HubSpot's REST API.
// It abstracts API endpoint construction, versioning, and JSON response processing
// specific to the HubSpot Batch feature.
type Adapter struct {
	Client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	// Batch updating objects does not support manipulating associations.
	// associationsStrategy is used to create associations as a follow up.
	associationsStrategy *associations.Strategy
}

// NewAdapter creates a new batch Adapter configured to work with Hubspot's APIs.
func NewAdapter(
	hubspotCRMClient *common.HTTPClient, providerInfo *providers.ProviderInfo,
	associationsStrategy *associations.Strategy,
) *Adapter {
	shouldHandleError := func(response *http.Response) bool {
		// 2xx responses are normal.
		// 400 (Bad Request) and 409 (Conflict) are considered valid "soft failures"
		// because HubSpot returns structured error information for these.
		// Any other status (e.g., 404, 5xx) represents a provider or implementation error.
		allowedCodes := datautils.NewSet(http.StatusBadRequest, http.StatusConflict)

		return !httpkit.Status2xx(response.StatusCode) &&
			!allowedCodes.Has(response.StatusCode)
	}

	jsonHTTPClient := &common.JSONHTTPClient{
		HTTPClient: &common.HTTPClient{
			Client:            hubspotCRMClient.Client,       // same authentication as Hubspot CRM
			ErrorHandler:      hubspotCRMClient.ErrorHandler, // same understanding of error format as Hubspot CRM
			ShouldHandleError: shouldHandleError,             // differs from CRM
		},
	}

	return &Adapter{
		Client:               jsonHTTPClient,
		providerInfo:         providerInfo,
		associationsStrategy: associationsStrategy,
	}
}

// getCreateURL builds the HubSpot batch create endpoint for the given object type.
//
// nolint:lll
// Contacts example: https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/batch/create-contacts
func (a *Adapter) getCreateURL(objectName common.ObjectName) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.providerInfo.BaseURL, "crm", "objects", core.APIVersion2026March, objectName.String(), "batch/create")
}

// getUpdateURL builds the HubSpot batch update endpoint for the given object type.
//
// nolint:lll
// Contacts example: https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/batch/update-contacts
func (a *Adapter) getUpdateURL(objectName common.ObjectName) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.providerInfo.BaseURL, "crm", "objects", core.APIVersion2026March, objectName.String(), "batch/update")
}

// getReadURL builds the HubSpot batch read endpoint for the given object type.
func (a *Adapter) getReadURL(objectName common.ObjectName) (*urlbuilder.URL, error) {
	if core.ObjectMarketingCampaigns == objectName {
		// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/batch/get-campaigns
		return urlbuilder.New(a.providerInfo.BaseURL, "marketing/campaigns", core.APIVersion2026March, "batch/read")
	}

	if core.ObjectContacts == objectName {
		// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/batch/get-contacts
		return urlbuilder.New(a.providerInfo.BaseURL, "crm/objects", core.APIVersion2026March, "contacts", "batch/read")
	}

	return nil, common.ErrObjectNotSupported
}
