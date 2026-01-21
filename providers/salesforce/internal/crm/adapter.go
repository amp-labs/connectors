package crm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/batch"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

// Adapter handles CRUD operations (at the moment: delete only) against Salesforce's REST API.
// It abstracts API endpoint construction, versioning, and JSON response processing
// specific to the Salesforce CRUD feature.
type Adapter struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.Deleter

	// CRM module sub-adapters.
	// These delegate specialized subsets of CRM functionality to keep Connector modular and prevent code bloat.
	customAdapter *metadata.Adapter // used for connectors.UpsertMetadataConnector capabilities.
	batchAdapter  *batch.Adapter    // used for connectors.BatchWriteConnector capabilities.
}

// NewAdapter creates a new crm Adapter configured to work with Salesforce's APIs.
func NewAdapter(params *common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Salesforce, *params, constructor)
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
			ErrorHandler:  core.NewErrorHandler().Handle,
		},
	)

	// Delegate selected CRM functionality to internal adapters to
	// prevent this package from growing too large. These adapters
	// effectively "inline" specialized responsibilities while sharing
	// the same HTTP and module context.
	adapter.customAdapter = metadata.NewAdapter(adapter.HTTPClient(), adapter.JSONHTTPClient(), adapter.ModuleInfo())
	adapter.batchAdapter = batch.NewAdapter(adapter.HTTPClient(), adapter.ModuleInfo())

	return adapter, nil
}

func (a Adapter) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	// Delegated.
	return a.customAdapter.UpsertMetadata(ctx, params)
}

func (a Adapter) BatchWrite(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error) {
	// Delegated.
	return a.batchAdapter.BatchWrite(ctx, params)
}

// Gateway access to URLs.
func (a Adapter) getModuleURL() string {
	return a.ModuleInfo().BaseURL
}

// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_sobject_retrieve_delete.htm
func (a Adapter) getDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), core.URISobjects, objectName, recordID)
}
