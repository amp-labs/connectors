package subscribe

import (
	"context"
	"time"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/microsoft"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// microsoftConfig is the provider configuration for Microsoft subscriptions.
//
// Microsoft subscribes via API, so a buildRequestFn is required for WebhookURL.
// Message verification is supported by the connector.
// Maintenance runs every 5h.
//
// No special setup is required for Registration/PostProcess.
var microsoftConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getMicrosoftRequest,
	},
	Maintenance: MaintenanceConfig{
		// The average lifetime limit for supported objects is 6 hours.
		// Connector creates subscriptions that last for 5:55, almost 6 hours,
		// which satisfies the subscription payload requirements.
		//
		// We should still trigger refresh before the SubscriptionResource expires.
		// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#subscription-lifetime
		renewalInterval: 5 * time.Hour, // nolint:mnd
	},
	Verification: VerificationConfig{
		verifierConnector: &microsoft.Connector{},
		bypassed:          true,
	},
}

func getMicrosoftRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &microsoft.SubscriptionRequest{
		WebhookURL: webhookURL,
	}, nil
}
