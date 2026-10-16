package jobber

import (
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	// Provides typed EmptySubscriptionParams/EmptySubscriptionResult for the
	// SubscribeConnector interface.
	components.SubscriptionInputOutput[SubscriptionRequest, SubscriptionResult]
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// Create base connector with provider info
	return components.Initialize(providers.Jobber, params, constructor)
}

// Provider returns the provider this connector talks to. It is declared here rather than
// inherited from the embedded base because the subscribe registry registers a zero-value
// Connector as its webhook verifier, which has no base: the base's Provider has a value
// receiver, so the promoted call would dereference that nil pointer and panic.
func (c *Connector) Provider() providers.Provider {
	return providers.Jobber
}

// String returns a human-readable identifier for this connector. Declared for the same reason as
// Provider: the zero-value verifier connector has no base to delegate to.
//
// Unlike Provider above, this still delegates instead of returning a constant: the base renders
// "<provider>.Connector[<module>]", so collapsing it would drop the module from existing log
// output. The constant is only the fallback for when there is no base to ask.
func (c *Connector) String() string {
	if c == nil || c.Connector == nil {
		return c.Provider() + ".Connector"
	}

	return c.Connector.String()
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	// Set the metadata provider for the connector
	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
		},
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler: interpreter.ErrorHandler{
				JSON: interpreter.NewFaultyResponder(errorFormats, nil),
			}.Handle,
		},
	)

	return connector, nil
}
