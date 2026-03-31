package restlet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

var ErrMissingMetadata = errors.New("missing required metadata")

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

// NewAdapter creates a RESTlet adapter. It reads scriptId and deployId
// from params.Metadata to construct the RESTlet URL.
func NewAdapter(params common.ConnectorParams) (*Adapter, error) { //nolint:funlen
	return newAdapter(providers.Netsuite, params)
}

// NewAdapterForProvider creates an adapter for a given provider (e.g. NetsuiteM2M).
func NewAdapterForProvider(provider providers.Provider, params common.ConnectorParams) (*Adapter, error) {
	return newAdapter(provider, params)
}

func newAdapter(provider providers.Provider, params common.ConnectorParams) (*Adapter, error) { //nolint:funlen
	return components.Initialize(provider, params, func(base *components.Connector) (*Adapter, error) {

		var scriptURL, scriptId, deployId string
		var ok bool

		// If scriptURL is provided, use it to build the restlet URL.
		scriptURL, ok = params.Metadata["scriptURL"]
		if !ok {
			// If scriptURL is not provided, use scriptId and deployId to build the restlet URL.
			scriptId, ok = params.Metadata["scriptId"]
			if !ok || scriptId == "" {
				return nil, fmt.Errorf("%w: scriptId", ErrMissingMetadata)
			}

			deployId, ok = params.Metadata["deployId"]
			if !ok || deployId == "" {
				return nil, fmt.Errorf("%w: deployId", ErrMissingMetadata)
			}
		}

		baseURL := base.ModuleInfo().BaseURL

		restletURL, err := buildRestletURL(baseURL, scriptURL, scriptId, deployId)
		if err != nil {
			return nil, fmt.Errorf("failed to build restlet URL: %w", err)
		}

		adapter := &Adapter{
			Connector:  base,
			restletURL: restletURL,
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

func buildRestletURL(baseURL, scriptURL, scriptId, deployId string) (string, error) {
	rurl, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if scriptURL != "" {
		ref, err := url.Parse(scriptURL)
		if err != nil {
			return "", err
		}
		return rurl.ResolveReference(ref).String(), nil
	}

	rurl.Path = "/app/site/hosting/restlet.nl"

	q := rurl.Query()
	q.Set("script", scriptId)
	q.Set("deploy", deployId)
	rurl.RawQuery = q.Encode()

	return rurl.String(), nil
}
