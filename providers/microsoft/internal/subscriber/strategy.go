package subscriber

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
	components.SubscriptionInputOutput[Input, Output]

	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	// Dependent services.
	batchStrategy *batch.Strategy
	clock         components.Clock
}

// NewStrategy constructs a Strategy with required dependencies.
func NewStrategy(
	client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, batchStrategy *batch.Strategy,
) *Strategy {
	return &Strategy{
		client:        client,
		providerInfo:  providerInfo,
		batchStrategy: batchStrategy,
		clock:         components.NewRealClock(),
	}
}

// Input defines the subscription request payload for Microsoft Graph.
type Input struct {
	// WebhookURL is the target URL where messages will be sent.
	WebhookURL string `json:"notificationUrl"`
}

// Output represents the subscription result payload.
//
// Currently empty, as Microsoft does not return structured data
// that needs to be captured for this connector.
type Output struct{}

// ObjectName represents a Microsoft resource type used for subscriptions.
type ObjectName = common.ObjectName

// getSubscriptionURL builds the Microsoft Graph endpoint for subscription operations.
//
// Docs:
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#subscription-request
func (s Strategy) getSubscriptionURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, apiVersion, "subscriptions")
}

// constructTestStrategy creates a Strategy configured for unit testing.
//
// It uses a mock HTTP client and overrides the base URL to point to a test server.
// A fixed clock is injected to ensure deterministic behavior in tests.
func constructTestStrategy(serverURL string) (*Strategy, error) {
	transport, err := components.NewTransport(providers.Microsoft, common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestMockServerBaseURL(serverURL)

	client := transport.JSONHTTPClient()
	info := transport.ProviderInfo()
	strategy := NewStrategy(client, info, batch.NewStrategy(client, info))
	strategy.clock = components.NewFixedClock(time.Date(2026, 3, 4, 5, 0, 0, 0, time.UTC))

	return strategy, nil
}
