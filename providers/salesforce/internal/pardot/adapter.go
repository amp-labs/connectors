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

	connector := funcName(adapter)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return adapter, nil
}

func funcName(adapter *Adapter) *Adapter {
	connector := adapter

	return connector
}

func (a *Adapter) getModuleURL() string {
	return a.ModuleInfo().BaseURL
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.getModuleURL(), "api/v5/objects", objectName)
}
