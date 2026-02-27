package batch

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

// Adapter handles batched record operations (create/update) against HubSpot's REST API.
// It abstracts API endpoint construction, versioning, and JSON response processing
// specific to the HubSpot Batch feature.
type Adapter struct {
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

// NewAdapter creates a new batch Adapter configured to work with Hubspot's APIs.
func NewAdapter(hubspotCRMClient *common.HTTPClient, moduleInfo *providers.ModuleInfo) *Adapter {
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
		Client:     jsonHTTPClient,
		moduleInfo: moduleInfo,
	}
}

func (a *Adapter) getModuleURL() string {
	return a.moduleInfo.BaseURL
}

// getCreateURL builds the HubSpot batch create endpoint for the given object type.
//
// nolint:lll
// Contacts example: https://developers.hubspot.com/docs/api-reference/crm-contacts-v3/batch/post-crm-v3-objects-contacts-batch-create
func (a *Adapter) getCreateURL(objectName common.ObjectName) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), core.APIVersion3, "objects", objectName.String(), "batch/create")
}

// getUpdateURL builds the HubSpot batch update endpoint for the given object type.
//
// nolint:lll
// Contacts example: https://developers.hubspot.com/docs/api-reference/crm-contacts-v3/batch/post-crm-v3-objects-contacts-batch-update
func (a *Adapter) getUpdateURL(objectName common.ObjectName) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), core.APIVersion3, "objects", objectName.String(), "batch/update")
}
