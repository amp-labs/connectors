package subscriber

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const (
	apiVersion     = "v4_6_release/apis/3.0"
	messageVersion = "3.0.0"
)

// Strategy implements subscription lifecycle operations.
//
// It embeds SubscriptionInputOutput to provide type-safe handling of
// subscription input/output while conforming to the non-generic connectors.SubscribeConnector interface.
type Strategy struct {
	components.SubscriptionInputOutput[Request, Result]

	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	clientID string
}

// NewStrategy constructs a Strategy with required dependencies.
func NewStrategy(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, clientID string) *Strategy {
	return &Strategy{
		client:       client,
		providerInfo: providerInfo,
		clientID:     clientID,
	}
}

// Request defines the subscription request payload for the webhook provider.
type Request struct {
	// WebhookURL is the endpoint where webhook messages will be delivered.
	WebhookURL string
}

// Result represents the subscription result payload.
type Result struct {
	// ObjectWebhooks maps object names to their associated webhook resources.
	ObjectWebhooks map[common.ObjectName]SubscriptionResource
}

// getSubscriptionURL builds the Microsoft Graph endpoint for subscription operations.
//
// Docs:
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#subscription-request
func (s Strategy) getSubscriptionURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, apiVersion, "system/callbacks")
}

func (s Strategy) clientIdHeader() common.Header {
	return common.Header{
		Key:   "ClientId",
		Value: s.clientID,
	}
}
