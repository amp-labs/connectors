package subscribe

import (
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
)

// mockAttioConfig is the subscribe-config bundle for the mockattio test provider. It mirrors
// attioConfig — the same subscribe-time request builder and WatchFieldsAuto quirk, and its
// object-shaped webhook payloads are fanned out by Attio's real CollapsedSubscriptionEvent in
// GetObjectTypeSubscribeEventsList — while webhook verification is bypassed (the real config's
// stored-secret HMAC needs subscription results the mock does not create) and the verifier
// connector serves canned records without HTTP. The connector.New factory additionally wires
// the mock connector with an object-name resolver answering record.* events' id.object_id from
// the store's seeded index, standing in for Attio's API-backed GetObjectNameFromEvent.
var mockAttioConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn:          getAttioRequest,
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		bypassed:          true,
		verifierConnector: mocksub.NewConnector(providers.MockAttio),
	},
}
