package connectwise

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise/internal/batch"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
	"github.com/amp-labs/connectors/providers/connectwise/internal/subscriber"
	"github.com/amp-labs/connectors/providers/connectwise/internal/webhook"
)

const apiVersion = "v4_6_release/apis/3.0"

// Type Exports.
type (
	SubscriptionEvent          = webhook.Event
	CollapsedSubscriptionEvent = webhook.CollapsedSubscriptionEvent
	SubscriptionRequest        = subscriber.Request
	SubscriptionResponse       = subscriber.Result
	SubscriptionResource       = subscriber.SubscriptionResource
)

type subscribeStrategy = subscriber.Strategy

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
	// TODO must use webhook.Verifier instead of webhook.NoopVerifier
	*webhook.NoopVerifier
	*subscribeStrategy

	batchAdapter *batch.Adapter // used for connectors.BatchRecordReaderConnector capabilities.

	clientID string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.ConnectWise, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	clientID := params.Metadata["clientId"]
	connector := &Connector{
		Connector:            base,
		ExpectedMetadataKeys: []string{"clientId"},
		clientID:             clientID,
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle
	connector.SetErrorHandler(errorHandler)

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		common.ModuleRoot,
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

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

	connector.batchAdapter = batch.NewAdapter(connector.JSONHTTPClient(), connector.ProviderInfo(), clientID)
	connector.NoopVerifier = webhook.NewVerifier(connector.JSONHTTPClient(), connector.ProviderInfo(), clientID)
	connector.subscribeStrategy = subscriber.NewStrategy(connector.JSONHTTPClient(), connector.ProviderInfo(), clientID)

	return connector, nil
}

// Provider returns the provider this connector talks to. It is declared here rather than
// inherited from the embedded base because the subscribe registry registers a zero-value
// Connector as its webhook verifier, which has no base: the base's Provider has a value
// receiver, so the promoted call would dereference that nil pointer and panic.
func (c *Connector) Provider() providers.Provider {
	return providers.ConnectWise
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

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ModuleInfo().BaseURL, apiVersion, objectPath)
}

func (c *Connector) getCommunicationTypesURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ModuleInfo().BaseURL, apiVersion, "company/communicationTypes")
}

func (c *Connector) clientIdHeader() common.Header {
	return common.Header{
		Key:   "ClientId",
		Value: c.clientID,
	}
}
