package subscribe

import (
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
)

// mockSalesloftConfig is the subscribe-config bundle for the mocksalesloft test provider. It
// mirrors salesloftConfig — the same subscribe-time request builder and WatchFieldsAuto quirk,
// so regression tests exercise the same routing behavior — but declares webhook verification
// bypassed (happy-path tests carry no signatures) and a mocksub verifier connector serving
// canned records without HTTP. Object-shaped webhook payloads are dispatched to Salesloft's
// real CollapsedSubscriptionEvent in GetObjectTypeSubscribeEventsList, so event parsing and
// classification run the real Salesloft code.
var mockSalesloftConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn:          getSalesloftRequest,
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		bypassed:          true,
		verifierConnector: mocksub.NewConnector(providers.MockSalesloft),
	},
}
