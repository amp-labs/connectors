package crm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

// Adapter handles CRUD operations (at the moment: delete only) against HubSpot's REST API.
// It abstracts API endpoint construction, versioning, and JSON response processing
// specific to the HubSpot CRUD feature.
type Adapter struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.Deleter
}

// NewAdapter creates a new crm Adapter configured to work with Hubspot's APIs.
func NewAdapter(params *common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Hubspot, *params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{Connector: base}

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  core.InterpretJSONError,
		},
	)

	return adapter, nil
}

func (a *Adapter) getModuleURL() string {
	return a.ModuleInfo().BaseURL
}

func (a *Adapter) getDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), core.APIVersion3, "objects", objectName, recordID)
}
