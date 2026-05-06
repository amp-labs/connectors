package marketing

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/shared"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

// Adapter handles CRUD operations against HubSpot's Marketing Hub product.
type Adapter struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

// NewAdapter creates a new Marketing Adapter configured to work with Hubspot's Marketing APIs.
func NewAdapter(params *common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Hubspot, *params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{Connector: base}

	adapter.SchemaProvider = schema.NewOpenAPISchemaProvider(adapter.ProviderContext.Module(), Schemas)

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler:  shared.InterpretJSONError,
		},
	)

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  shared.InterpretJSONError,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  shared.InterpretJSONError,
		},
	)

	return adapter, nil
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	path, err := Schemas.FindURLPath(a.Module(), objectName)
	if err != nil {
		return nil, common.ErrOperationNotSupportedForObject
	}

	return urlbuilder.New(a.ModuleInfo().BaseURL, path, shared.APIVersion2026March)
}

func constructTestAdapter(serverURL string) (*Adapter, error) {
	adapter, err := NewAdapter(
		&common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
			Module:              providers.ModuleHubspotMarketing,
		},
	)
	if err != nil {
		return nil, err
	}

	adapter.SetUnitTestMockServerBaseURL(serverURL)

	return adapter, nil
}
