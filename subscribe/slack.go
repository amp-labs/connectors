package subscribe

import (
	"github.com/amp-labs/connectors/providers/slack"
)

// slackConfig is the per-provider subscribe-config bundle for Slack.
//
// Slack is not capable of managing subscriptions programmatically.
var slackConfig = ProviderConfig{
	Verification: VerificationConfig{
		verifierConnector: &slack.Connector{},
	},
}
