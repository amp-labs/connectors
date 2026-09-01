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

func getSlackVerificationParams(
	ctx context.Context,
	deps deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil || req.Installation == nil {
		return nil, errInstallationNotFound
	}

	if deps.Subscriptions == nil {
		return nil, errSubscriptionListerNotConfigured
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
