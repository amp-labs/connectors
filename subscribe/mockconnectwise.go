package subscribe

import (
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
)

// mockConnectWiseConfig is the subscribe-config bundle for the mockconnectwise test provider.
// It mirrors connectWiseConfig — the same subscribe-time request builder, the WatchFieldsAuto
// quirk, and bypassed verification (the real connector's verification needs an authenticated
// client, so ConnectWise itself is bypassed too) — while the verifier connector serves canned
// records without HTTP. Object-shaped webhook payloads are dispatched to ConnectWise's real
// CollapsedSubscriptionEvent in GetObjectTypeSubscribeEventsList, so event parsing — including
// the inline-record Entity extraction (SubscriptionEventWithRecord) — runs the real
// ConnectWise code.
var mockConnectWiseConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn:          getConnectWiseRequest,
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		bypassed:          true,
		verifierConnector: mocksub.NewConnector(providers.MockConnectWise),
	},
}
