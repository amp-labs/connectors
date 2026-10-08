package subscribe

import (
	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot"
)

// mockHubspotConfig is the subscribe-config bundle for the mockhubspot test provider. It
// mirrors hubspotConfig — look-up-only (no subscribe/registration declarations), verification
// bypassed, and crucially the same event caster, so regression tests cast webhook payloads
// through HubSpot's real SubscriptionEvent implementation — while the verifier connector
// serves canned records without HTTP. The verification-params builder is omitted: it feeds
// signature verification, which the mock bypasses.
var mockHubspotConfig = ProviderConfig{
	Verification: VerificationConfig{
		bypassed:          true,
		verifierConnector: mocksub.NewConnector(providers.MockHubspot),
		eventCaster:       CastSubscriptionEvents[hubspot.SubscriptionEvent],
	},
}
