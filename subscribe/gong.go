package subscribe

import (
	"github.com/amp-labs/connectors/providers/gong"
)

// gongConfig is the per-provider subscribe-config bundle for Gong. Gong is look-up-only, so it has
// no subscribe/registration declarations; it carries only webhook-verification data — a verifier
// connector and a signature-verification bypass (verification not yet implemented).
var gongConfig = ProviderConfig{
	Verification: VerificationConfig{
		verifierConnector: &gong.Connector{},
		bypassed:          true,
	},
}
