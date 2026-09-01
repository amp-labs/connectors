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

func NewUserConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.SlackUserScope, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	authMetadata := NewAuthMetadataVars(params.Metadata)
	// Signing Secret is used by the event message verifier.
	// If the value is empty then all messages will be marked as invalid.
	signingSecret := params.Metadata["webhookSigningSecret"]
	verifier := webhook.NewVerifier(signingSecret)

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
