package restlet

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

// Adapter implements Read, Write, Delete, and ListObjectMetadata by
// talking to a NetSuite RESTlet script over a single POST endpoint.
type Adapter struct {
	*components.Connector
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	// restletURL is the fully-qualified URL to the RESTlet endpoint,
	// including script and deploy query params.
	restletURL string
}

// NewAdapter creates a RESTlet adapter. The RESTlet URL comes from the module's
// BaseURL, which includes script and deploy query params resolved via catalog substitution.
func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Netsuite, params, func(base *components.Connector) (*Adapter, error) {
		// The module BaseURL already contains the full RESTlet path with script and deploy query params,
		// resolved from metadata via catalog variable substitution.
		adapter := &Adapter{
			Connector:  base,
			restletURL: base.ModuleInfo().BaseURL,
		}

		registry := components.NewEmptyEndpointRegistry()
		httpClient := adapter.HTTPClient().Client

		adapter.SchemaProvider = schema.NewObjectSchemaProvider(
			httpClient,
			schema.FetchModeSerial,
			operations.SingleObjectMetadataHandlers{
				BuildRequest:  adapter.buildObjectMetadataRequest,
				ParseResponse: adapter.parseObjectMetadataResponse,
				ErrorHandler:  common.InterpretError,
			},
		)

		adapter.Reader = reader.NewHTTPReader(
			httpClient,
			registry,
			adapter.ProviderContext.Module(),
			operations.ReadHandlers{
				BuildRequest:  adapter.buildReadRequest,
				ParseResponse: adapter.parseReadResponse,
				ErrorHandler:  common.InterpretError,
			},
		)

		adapter.Writer = writer.NewHTTPWriter(
			httpClient,
			registry,
			adapter.ProviderContext.Module(),
			operations.WriteHandlers{
				BuildRequest:  adapter.buildWriteRequest,
				ParseResponse: adapter.parseWriteResponse,
				ErrorHandler:  common.InterpretError,
			},
		)

		adapter.Deleter = deleter.NewHTTPDeleter(
			httpClient,
			registry,
			adapter.ProviderContext.Module(),
			operations.DeleteHandlers{
				BuildRequest:  adapter.buildDeleteRequest,
				ParseResponse: adapter.parseDeleteResponse,
				ErrorHandler:  common.InterpretError,
			},
		)

		return adapter, nil
	})
}
