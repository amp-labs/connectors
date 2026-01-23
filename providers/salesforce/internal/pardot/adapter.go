package pardot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce/internal/pardot/metadata"
)

const MetadataKeyBusinessUnitID = "businessUnitId"

type Adapter struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireMetadata

	// Supported operations
	components.Reader
	components.Writer
	components.Deleter

	// Features.
	metadataStrategy *metadata.Strategy

	// Variables.
	businessUnitID string
}

// NewAdapter creates a new pardot Adapter configured to work with Salesforce's Account Management APIs.
func NewAdapter(params *common.ConnectorParams) (*Adapter, error) {
	adapter, err := components.Initialize(providers.Salesforce, *params, constructor)
	if err != nil {
		return nil, err
	}

	adapter.businessUnitID = params.Metadata[MetadataKeyBusinessUnitID]

	return adapter, nil
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{MetadataKeyBusinessUnitID},
		},
	}

	var err error

	adapter.metadataStrategy, err = metadata.NewStrategy(base)
	if err != nil {
		return nil, err
	}

	errorHandler := errorHandlerFunc

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return adapter, nil
}

func (a *Adapter) getModuleURL() string {
	return a.ModuleInfo().BaseURL
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), "api/v5/objects", objectName)
}
