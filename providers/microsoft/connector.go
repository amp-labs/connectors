package microsoft

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
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
	"github.com/amp-labs/connectors/providers/microsoft/internal/metadata"
	"github.com/amp-labs/connectors/providers/microsoft/internal/subscriber"
	"github.com/amp-labs/connectors/providers/microsoft/internal/webhook"
)

const apiVersion = "v1.0"

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
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// Supported operations
	components.Reader
	components.Writer
	components.Deleter
	*webhook.NoopVerifier
	*subscribeStrategy

	// Dependent services.
	batchStrategy *batch.Strategy
}

// NewConnector creates a new Microsoft connector. It defaults to the Microsoft
// provider; use NewConnectorForProvider for twin providers (e.g.
// MicrosoftAdminConsent) that share the same implementation but differ
// in auth scheme.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return NewConnectorForProvider(providers.Microsoft, params)
}

// NewConnectorForProvider creates a new Microsoft connector under the given
// provider name. This allows twin providers like MicrosoftAdminConsent
// to reuse the same connector implementation with a different auth
// configuration.
func NewConnectorForProvider(provider providers.Provider, params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(provider, params, constructor)
}

// nolint:funlen
func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector:     base,
		batchStrategy: batch.NewStrategy(base.JSONHTTPClient(), base.ProviderInfo()),
	}

	// DirectFaultyResponder (vs. the default FaultyResponder) gives the
	// callback access to the raw *http.Response, including headers.
	// handleErrorResponse needs WWW-Authenticate to detect CAE / step-up
	// claim challenges on 401s; see providers/microsoft/errors.go for the
	// classification logic and known limitations.
	// Fallback catches Graph's bodyless 500s, which arrive with no Content-Type.
	errorHandler := interpreter.ErrorHandler{
		JSON:     interpreter.DirectFaultyResponder{Callback: handleErrorResponse},
		Fallback: interpreter.DirectFaultyResponder{Callback: handleNonJSONErrorResponse},
	}.Handle

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
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

	connector.NoopVerifier = webhook.NewVerifier(connector.JSONHTTPClient(), connector.ProviderInfo())
	connector.batchStrategy = batch.NewStrategy(base.JSONHTTPClient(), base.ProviderInfo())
	connector.subscribeStrategy = subscriber.NewStrategy(
		base.JSONHTTPClient(), base.ProviderInfo(), connector.batchStrategy)

	return connector, nil
}

func (c *Connector) getReadUrl(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
}

func (c *Connector) getWriteUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectDrafts:
		// https://learn.microsoft.com/en-us/graph/api/user-list-messages
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/messages")
	case virtualObjectSentMessages:
		// https://learn.microsoft.com/en-us/graph/api/user-sendmail
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/sendMail")
	case objectNameMessages:
		// Writing directly to messages is hidden.
		// Customers must use virtual objects: drafts and sentMessages.
		return nil, common.ErrObjectNotSupported
	}

	return c.getReadUrl(objectName)
}

func (c *Connector) getDeleteUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectDrafts:
		// https://learn.microsoft.com/en-us/graph/api/user-list-messages
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/messages")
	case virtualObjectSentMessages:
		return nil, common.ErrObjectNotSupported
	case objectNameMessages:
		return nil, common.ErrObjectNotSupported
	}

	return c.getReadUrl(objectName)
}
