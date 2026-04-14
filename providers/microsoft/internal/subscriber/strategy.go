package subscriber

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

const apiVersion = "v1.0"

type Strategy struct {
	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo

	// Dependent services.
	batchStrategy *batch.Strategy

	components.SubscriptionInputOutput[Input, Output]
}

func NewStrategy(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo) *Strategy {
	return &Strategy{
		client:        client,
		providerInfo:  providerInfo,
		batchStrategy: batch.NewStrategy(client, providerInfo),
	}
}

type Input struct {
	// WebhookURL is the target URL where messages will be sent.
	WebhookURL string `json:"notificationUrl"`
}

// Output TODO describe what is it.
type Output map[string]SubscriptionResource

// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#subscription-request
func (s Strategy) getCreateSubscriptionURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(s.providerInfo.BaseURL, apiVersion, "subscriptions")
}
