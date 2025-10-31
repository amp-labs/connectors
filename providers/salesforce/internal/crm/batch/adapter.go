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

// getCreateURL builds the endpoint for creating multiple records of the same object type.
//
// Object name is required as a suffix of the URL.
// Only one type of objects can be created at a time, this is by Salesforce API design.
//
// nolint:lll
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/dome_composite_sobject_tree_flat.htm
func (a *Adapter) getCreateURL(objectName common.ObjectName) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), restAPISuffix, "/composite/tree", objectName.String())
}

// getUpdateURL builds the endpoint for updating multiple records across one or more object types.
//
// Objects of multiple type can be created as part of one request, and it is not limited to one objectName
// and therefore no such argument is needed unlike the getCreateURL.
//
// nolint:lll
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_update.htm
func (a *Adapter) getUpdateURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), restAPISuffix, "/composite/sobjects")
}
