package subscribe

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// slackConfig is the per-provider subscribe-config bundle for Slack.
//
// Slack is not capable of managing subscriptions programmatically.
var slackConfig = ProviderConfig{
	Verification: VerificationConfig{
		paramsFn:          getSlackVerificationParams,
		verifierConnector: &slack.Connector{},
	},
}

// getSlackVerificationParams resolves the Slack signing secret from the provider app's metadata
// (ProviderApp.metadata.providerParams). Like Hubspot, Slack webhooks are integration-scoped —
// the Events API request URL is configured once per Slack app, so no installation exists at
// verification time and req.Installation must not be required here. An empty secret is passed
// through; the verifier rejects it with a clear ErrMissingProviderParam.
func getSlackVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil {
		return nil, errNilVerificationRequest
	}

	secret := ""

	if req.ProviderApp != nil && req.ProviderApp.Metadata != nil && req.ProviderApp.Metadata.ProviderParams != nil {
		params := *req.ProviderApp.Metadata.ProviderParams
		secret = params[providers.ProviderParamSlackSigningSecret]
	}

	return &common.VerificationParams{
		Param: &slack.VerificationParams{SigningSecret: secret},
	}, nil
}
