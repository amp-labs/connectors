package subscribe

import (
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
)

// mockSalesforceConfig is the subscribe-config bundle for the mocksalesforce test provider.
// Its object-shaped webhook payloads take Salesforce's real receive path in
// GetObjectTypeSubscribeEventsList — the AWS EventBridge envelope unwrap and the
// per-recordIds fan-out into salesforce.SubscriptionEvent values — so regression tests parse
// and classify events through the real Salesforce code.
//
// Verification is declared bypassed, unlike the real config: Salesforce's non-bypassed path
// makes the server fetch a ProviderApp row even though its VerifyWebhookMessage accepts
// everything, and the mock skips that requirement. The Registration/Subscription builders are
// omitted: they feed subscribe-creation-time AWS EventBridge / CDC-optimization machinery with
// no mock counterpart, which the receive path under test never invokes.
var mockSalesforceConfig = ProviderConfig{
	Verification: VerificationConfig{
		bypassed:          true,
		verifierConnector: mocksub.NewConnector(providers.MockSalesforce),
	},
}
