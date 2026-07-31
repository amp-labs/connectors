package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// connectWiseConfig is the per-provider subscribe-config bundle for ConnectWise.
//
// ConnectWise subscribes via API and builds a subscribe-time request payload (webhook URL),
// so a buildRequestFn is declared. ConnectWise requires WatchFieldsAuto="all" for subscribe
// update events.
// Message Verification is supported by the connector, but requires AuthenticatedClient,
// therefore bypassed.
//
// No special setup is required for Registration/Maintenance/PostProcess.
var connectWiseConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn:          getConnectWiseRequest,
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		verifierConnector: &connectwise.Connector{},
		bypassed:          true,
	},
}

func getConnectWiseRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &connectwise.SubscriptionRequest{
		WebhookURL: webhookURL,
	}, nil
}
