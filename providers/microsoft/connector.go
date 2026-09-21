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

// Provider returns the provider this connector talks to.
//
// It is declared here rather than inherited from the embedded *components.Connector because the
// subscribe registry uses a zero-value Connector as its webhook verifier, which has no base. The
// base's Provider has a value receiver, so the promoted call would dereference that nil pointer
// and panic before any verification ran.
//
// This delegates rather than returning a constant: NewConnectorForProvider builds this connector
// under twin providers too (MicrosoftAdminConsent), so only the base knows which one this is.
func (c *Connector) Provider() providers.Provider {
	if c == nil || c.Connector == nil {
		return providers.Microsoft
	}

	return c.Connector.Provider()
}

// String returns a human-readable identifier for this connector. Declared for the same reason as
// Provider: the zero-value verifier connector has no base to delegate to.
func (c *Connector) String() string {
	if c == nil || c.Connector == nil {
		return c.Provider() + ".Connector"
	}

	return c.Connector.String()
}

// getCoreObjectReadUrl returns the Microsoft Graph URL used to read a core object.
//
// Core objects are resolved through the metadata schema registry. The object name
// must correspond to a schema registered under common.ModuleRoot.
//
// Unlike virtual objects, core objects use their metadata-defined URL path for
// read operations.
func (c *Connector) getCoreObjectReadUrl(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
}

// getReadUrl returns the Microsoft Graph URL used to read the specified object.
//
// Core objects are resolved through the metadata schema registry. Virtual mail
// objects are mapped to the corresponding mail-folder messages endpoint because
// they represent filtered views of the core message resource:
//
//   - virtualObjectDrafts maps to the Drafts folder.
//   - virtualObjectSentMessages maps to the Sent Items folder.
//   - virtualObjectInboxMessages maps to the Inbox folder.
//
// The core messages object itself is resolved through getCoreObjectReadUrl.
// Microsoft Graph documents listing messages in a mail folder at:
//
// https://learn.microsoft.com/en-us/graph/api/mailfolder-list-messages
func (c *Connector) getReadUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectDrafts:
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/mailFolders/drafts/messages")
	case virtualObjectSentMessages:
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/mailFolders/sentitems/messages")
	case virtualObjectInboxMessages:
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/mailFolders/inbox/messages")
	case objectNameMessages:
		// https://learn.microsoft.com/en-us/graph/api/user-list-messages
		fallthrough
	default:
		return c.getCoreObjectReadUrl(objectName)
	}
}

// getWriteUrl returns the Microsoft Graph URL used to write the specified object.
//
// Virtual objects do not necessarily have a one-to-one correspondence with a
// writable Microsoft Graph resource:
//
//   - Drafts are created through /me/messages.
//   - Inbox messages cannot be written to; sending a message uses /me/sendMail.
//   - Sent messages are read-only through this connector.
//   - The core messages object is intentionally not writable. Callers must use
//     the appropriate virtual object instead.
//
// For all other core objects, the read and write paths are identical and the
// path is resolved through the metadata schema registry.
func (c *Connector) getWriteUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectDrafts:
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/mailFolders/drafts/messages")
	case virtualObjectSentMessages:
		return nil, common.ErrOperationNotSupportedForObject
	case virtualObjectInboxMessages:
		// https://learn.microsoft.com/en-us/graph/api/user-sendmail
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/sendMail")
	case objectNameMessages:
		return nil, common.ErrOperationNotSupportedForObject
	default:
		return c.getCoreObjectReadUrl(objectName)
	}
}

// getDeleteUrl returns the Microsoft Graph URL used to delete the specified object.
//
// Drafts are deleted through the general messages' endpoint.
// Sent messages, inbox messages, and the core messages object are not deletable through this
// connector and return common.ErrObjectNotSupported.
//
// For all other core objects, the read and delete paths are identical and the
// path is resolved through the metadata schema registry.
func (c *Connector) getDeleteUrl(objectName string) (*urlbuilder.URL, error) {
	switch objectName {
	case virtualObjectDrafts:
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "/me/messages")
	case virtualObjectSentMessages:
		return nil, common.ErrOperationNotSupportedForObject
	case virtualObjectInboxMessages:
		return nil, common.ErrOperationNotSupportedForObject
	case objectNameMessages:
		return nil, common.ErrOperationNotSupportedForObject
	default:
		return c.getCoreObjectReadUrl(objectName)
	}
}
