package batch

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion    = "60.0"
	versionPrefix = "v"
	version       = versionPrefix + apiVersion
	restAPISuffix = "/services/data/" + version
)

// Adapter handles batched record operations (create/update) against Salesforce's REST API.
// It abstracts endpoint construction, versioning, and JSON response handling for the Batch feature.
type Adapter struct {
	Client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

// NewAdapter creates a new batch Adapter configured to work with Salesforce's composite APIs.
//
// Salesforce CRM client is used as a prototype a copy of which treats 400 BadRequest as permittable.
func NewAdapter(salesforceCRMClient *common.HTTPClient, moduleInfo *providers.ModuleInfo) *Adapter {
	shouldHandleError := func(response *http.Response) bool {
		// 2xx is allowed as well as 400 BadRequest.
		// All other responses need error handling.
		return !httpkit.Status2xx(response.StatusCode) && response.StatusCode != http.StatusBadRequest
	}

	jsonHTTPClient := &common.JSONHTTPClient{
		HTTPClient: &common.HTTPClient{
			Client:            salesforceCRMClient.Client,       // same authentication as Salesforce CRM
			ErrorHandler:      salesforceCRMClient.ErrorHandler, // same understanding of error format as Salesforce CRM
			ShouldHandleError: shouldHandleError,                // differs from CRM
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

// getCreateURL builds the endpoint for creating multiple records across one or more object types.
//
// nolint:lll
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_create.htm
func (a *Adapter) getCreateURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), restAPISuffix, "/composite/sobjects")
}

// getUpdateURL builds the endpoint for updating multiple records across one or more object types.
//
// nolint:lll
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_update.htm
func (a *Adapter) getUpdateURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), restAPISuffix, "/composite/sobjects")
}
