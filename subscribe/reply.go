package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/reply"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// replyConfig is the per-provider subscribe-config bundle for Reply. Reply subscribes via API
// (one webhook subscription per event type) and builds a subscribe-time request payload carrying
// the webhook URL, so a buildRequestFn is declared.
//
// Reply does not sign its deliveries: there is no secret, signature or event-name header, and the
// connector's VerifyWebhookMessage is a placeholder that accepts all, so verification is bypassed.
// An event caster unwraps its webhook payloads into typed SubscriptionEvents; each delivery
// carries a single event, fanned out by reply.CollapsedSubscriptionEvent (registered in
// GetObjectTypeSubscribeEventsList).
//
// No Registration (Reply has no shared resource to set up before subscribing) and no Maintenance
// (Reply documents no expiry for webhook subscriptions, so there is nothing to renew).
var replyConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getReplyRequest,
	},
	Verification: VerificationConfig{
		verifierConnector: &reply.Connector{},
		bypassed:          true,
		eventCaster:       CastSubscriptionEvents[reply.SubscriptionEvent],
	},
}

func getReplyRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &reply.SubscriptionRequest{
		WebhookURL: webhookURL,
	}, nil
}
