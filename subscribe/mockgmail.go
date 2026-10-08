package subscribe

import (
	"time"

	"github.com/amp-labs/connectors/mocksub"
	"github.com/amp-labs/connectors/providers"
	google "github.com/amp-labs/connectors/providers/google"
)

// mockGmailConfig is the subscribe-config bundle for the mockgmail test provider. It mirrors
// googleConfig (the Gmail module): verification bypassed — Gmail events reaching verification
// are synthetic republishes from the Gmail event workflow carrying no provider signature — and
// crucially the same event caster, so regression tests cast those synthetic payloads through
// Gmail's real SubscriptionEvent implementation. The maintenance renewal interval is mirrored
// for fidelity; the subscribe-time request builder (Gmail watch Pub/Sub topic resolution) is
// omitted, as the receive path under test never invokes it.
//
//nolint:mnd
var mockGmailConfig = ProviderConfig{
	Maintenance: MaintenanceConfig{
		renewalInterval: time.Hour * 24,
	},
	Verification: VerificationConfig{
		bypassed:          true,
		verifierConnector: mocksub.NewConnector(providers.MockGmail),
		eventCaster:       CastSubscriptionEvents[google.SubscriptionEvent],
	},
}
