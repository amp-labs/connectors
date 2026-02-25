package crm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/batch"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/custom"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/search"
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

	// CRM module sub-adapters
	// These delegate specialized subsets of CRM functionality to keep Connector modular and prevent code bloat.
	customAdapter      *custom.Adapter  // used for connectors.UpsertMetadataConnector capabilities.
	batchAdapter       *batch.Adapter   // used for connectors.BatchWriteConnector capabilities.
	searchStrategy     *search.Strategy // used for connectors.SearchConnector capabilities.
	AssociationsFiller associations.Filler
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

	adapter.SetErrorHandler(core.InterpretJSONError)
	adapter.customAdapter = custom.NewAdapter(adapter.JSONHTTPClient(), adapter.ModuleInfo())
	adapter.batchAdapter = batch.NewAdapter(adapter.HTTPClient(), adapter.ModuleInfo())
	adapter.AssociationsFiller = associations.NewStrategy(adapter.JSONHTTPClient(), adapter.ModuleInfo())
	adapter.searchStrategy = search.NewStrategy(
		adapter.JSONHTTPClient(), adapter.ModuleInfo(), adapter.AssociationsFiller,
	)

	return adapter, nil
}

func (a *Adapter) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	return a.customAdapter.UpsertMetadata(ctx, params)
}

func (a *Adapter) BatchWrite(
	ctx context.Context, params *common.BatchWriteParam,
) (*common.BatchWriteResult, error) {
	return a.batchAdapter.BatchWrite(ctx, params)
}

func (a *Adapter) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	return a.searchStrategy.Search(ctx, params)
}

func (a *Adapter) getModuleURL() string {
	return a.ModuleInfo().BaseURL
}

func (a *Adapter) getDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), core.APIVersion3, "objects", objectName, recordID)
}
