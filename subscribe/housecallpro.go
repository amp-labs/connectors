package subscribe

import (
	housecallpro "github.com/amp-labs/connectors/providers/housecallPro"
)

// housecallproConfig is the per-provider subscribe-config bundle for HousecallPro. HousecallPro is
// look-up-only, so it has no subscribe/registration declarations; it carries only webhook-
// verification data — a verifier connector and a signature-verification bypass (verification not
// yet implemented).
var housecallproConfig = ProviderConfig{
	Verification: VerificationConfig{
		verifierConnector: &housecallpro.Connector{},
		bypassed:          true,
	},
}
