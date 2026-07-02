package subscriber

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

const apiVersion = "v1.0"

// Strategy implements Microsoft subscription lifecycle operations.
//
// It embeds SubscriptionInputOutput to provide type-safe handling of
// subscription input/output while conforming to the non-generic connectors.SubscribeConnector interface.
//
// The strategy relies on:
//   - JSONHTTPClient for API communication
//   - batch.Strategy for batched operations
//   - Clock for time.Now. (useful in tests for deterministic outcomes)
type Strategy struct {
	components.SubscriptionInputOutput[Request, Result]

	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	// Dependent services.
	batchStrategy *batch.Strategy
}

// NewStrategy constructs a Strategy with required dependencies.
func NewStrategy(
	client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, batchStrategy *batch.Strategy,
) *Strategy {
	return &Strategy{
		client:        client,
		providerInfo:  providerInfo,
		batchStrategy: batchStrategy,
	}
}

// Request defines the subscription request payload for Microsoft Graph.
type Request struct {
	// WebhookURL is the target URL where messages will be sent.
	WebhookURL string `json:"notificationUrl"`
}

// Result represents the subscription result payload.
type Result struct {
	Subscriptions map[string]SubscriptionResource
}

// getSubscriptionURL builds the Microsoft Graph endpoint for subscription operations.
//
// Docs:
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#subscription-request
func (s Strategy) getSubscriptionURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, apiVersion, "subscriptions")
}
