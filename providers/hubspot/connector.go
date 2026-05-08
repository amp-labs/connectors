package hubspot

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/batch"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
	"github.com/amp-labs/connectors/providers/hubspot/internal/custom"
	"github.com/amp-labs/connectors/providers/hubspot/internal/search"
)

// Connector implements the HubSpot integration.
//
// HubSpot is modeled as a single connector.
// Unlike multi-module providers, all supported operations belong directly to Connector.
//
// Complex feature sets may still live in dedicated internal packages
// (for example batch, search, or subscriptions), but those packages are
// implementation details rather than top-level modules.
//
// The long-term direction is for Connector to own all operations directly,
// with reusable feature strategies embedded as needed.
type Connector struct {
	// Shared connector infrastructure.
	*components.Connector

	// Provides access to an authenticated client.
	common.RequireAuthenticatedClient

	// Operations
	components.Deleter

	// These delegate complex functionality to keep Connector modular and prevent code bloat.
	customAdapter      *custom.Adapter  // used for connectors.UpsertMetadataConnector capabilities.
	batchAdapter       *batch.Adapter   // used for connectors.BatchWriteConnector capabilities.
	searchStrategy     *search.Strategy // used for connectors.SearchConnector capabilities.
	associationsFiller associations.Filler
}

var _ connectors.WebhookVerifierConnector = &Connector{}

// NewConnector returns a new Hubspot connector.
// Hubspot connector still owns CRM functionality. Not every CRM feature is located under `crm` package.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Hubspot, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	// Note: error handler must return common.HTTPError.
	// Check method in the internal package "custom", method "readGroupName" which relies on error casting.
	connector.SetErrorHandler(core.InterpretJSONError)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  core.InterpretJSONError,
		},
	)

	connector.customAdapter = custom.NewAdapter(connector.JSONHTTPClient(), connector.ModuleInfo())
	associationsStrategy := associations.NewStrategy(
		connector.JSONHTTPClient(), connector.ModuleInfo(), connector.ProviderInfo(),
	)
	connector.associationsFiller = associationsStrategy
	connector.batchAdapter = batch.NewAdapter(connector.HTTPClient(), connector.ModuleInfo(), associationsStrategy)
	connector.searchStrategy = search.NewStrategy(
		connector.JSONHTTPClient(), connector.ModuleInfo(), connector.associationsFiller,
	)

	return connector, nil
}
