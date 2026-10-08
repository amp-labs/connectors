// Package acculynx provides a connector for the AccuLynx V2 API.
//
// API Documentation: https://apidocs.acculynx.com/
// Authentication: API key sent as a Bearer token in the Authorization header.
// Base URL: https://api.acculynx.com
package acculynx

import (
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/acculynx/metadata"
)

var (
	_ connectors.BatchRecordReaderConnector = &Connector{}
	_ connectors.WebhookVerifierConnector   = &Connector{}
	_ connectors.SubscribeConnector         = &Connector{}
)

// Connector is the AccuLynx connector.
type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

// NewConnector creates a new AccuLynx connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.AccuLynx, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(
		connector.ProviderContext.Module(),
		metadata.Schemas,
	)

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}

// Provider returns the provider this connector talks to. It is declared here rather than
// inherited from the embedded base because the subscribe registry registers a zero-value
// Connector as its webhook verifier, which has no base: the base's Provider has a value
// receiver, so the promoted call would dereference that nil pointer and panic.
func (c *Connector) Provider() providers.Provider {
	return providers.AccuLynx
}

// String returns a human-readable identifier for this connector. Declared for the same reason as
// Provider: the zero-value verifier connector has no base to delegate to.
func (c *Connector) String() string {
	if c == nil || c.Connector == nil {
		return c.Provider() + ".Connector"
	}

	return c.Connector.String()
}
