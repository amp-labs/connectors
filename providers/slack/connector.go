package slack

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
	"github.com/amp-labs/connectors/providers/slack/internal/webhook"
)

// Type Exports.
type (
	SubscriptionEvent          = webhook.Event
	CollapsedSubscriptionEvent = webhook.CollapsedSubscriptionEvent
	VerificationParams         = webhook.VerificationParams
)

var (
	_ connectors.ReadConnector              = (*Connector)(nil)
	_ connectors.WriteConnector             = (*Connector)(nil)
	_ connectors.BatchRecordReaderConnector = (*Connector)(nil)
	_ connectors.AuthMetadataConnector      = (*Connector)(nil)
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client & account
	common.RequireAuthenticatedClient
	common.PostAuthInfo

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
	*webhook.Verifier

	teamId string
}

func NewBotConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Slack, params, constructor)
}

// NewWebhookVerifierConnector returns a minimal Connector for webhook signature verification
// only. It carries a constructed webhook Verifier and nothing else — no auth client, no base
// connector — so it must not be used for API calls. The subscribe registry uses it as the
// shared verifierConnector: verification is stateless HMAC math, and a zero-value Connector's
// nil *Verifier embed would otherwise fail every verification (and, before the verifier's nil
// guard existed, panicked in the method-promotion wrapper).
func NewWebhookVerifierConnector() *Connector {
	return &Connector{Verifier: webhook.NewVerifier()}
}

func NewUserConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.SlackUserScope, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	authMetadata := NewAuthMetadataVars(params.Metadata)
	verifier := webhook.NewVerifier()

	connector := &Connector{
		Connector: base,
		Verifier:  verifier,
		teamId:    authMetadata.TeamId,
	}

	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
		},
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

// Provider returns the provider this connector talks to.
//
// It is declared here rather than inherited from the embedded *components.Connector because
// NewWebhookVerifierConnector builds a Connector with no base. The base's Provider has a value
// receiver, so the promoted call dereferences that nil pointer and panics — which is what the
// server hit when it wrapped the verifier connector for metrics, before any verification ran.
//
// This delegates rather than returning a constant: a Slack connector is built under two different
// providers — NewBotConnector under Slack and NewUserConnector under SlackUserScope — so only the
// base knows which one this is. Slack is the right answer for the baseless verifier connector,
// which the registry shares between both.
func (c *Connector) Provider() providers.Provider {
	if c == nil || c.Connector == nil {
		return providers.Slack
	}

	return c.Connector.Provider()
}

// String returns a human-readable identifier for this connector. Declared for the same reason as
// Provider: the verifier connector has no base to delegate to.
func (c *Connector) String() string {
	if c == nil || c.Connector == nil {
		return c.Provider() + ".Connector"
	}

	return c.Connector.String()
}
