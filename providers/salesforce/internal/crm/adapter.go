package crm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/batch"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/custom"
)

// Adapter handles CRUD operations (at the moment: delete only) against Salesforce's REST API.
// It abstracts API endpoint construction, versioning, and JSON response processing
// specific to the Salesforce CRUD feature.
type Adapter struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// CRM module sub-adapters.
	// These delegate specialized subsets of CRM functionality to keep Connector modular and prevent code bloat.
	customAdapter *custom.Adapter // used for connectors.UpsertMetadataConnector capabilities.
	batchAdapter  *batch.Adapter  // used for connectors.BatchWriteConnector capabilities.
}

// NewAdapter creates a new crm Adapter configured to work with Salesforce's APIs.
func NewAdapter(params *common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Salesforce, *params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{Connector: base}

	// Delegate selected CRM functionality to internal adapters to
	// prevent this package from growing too large. These adapters
	// effectively "inline" specialized responsibilities while sharing
	// the same HTTP and module context.
	adapter.customAdapter = custom.NewAdapter(adapter.HTTPClient(), adapter.JSONHTTPClient(), adapter.ModuleInfo())
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
