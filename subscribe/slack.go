package subscribe

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// errSlackSigningSecretNotConfigured is returned when the provider app has no Slack signing
// secret in its metadata. The builder sets it in the Ampersand dashboard (Provider Apps > Slack >
// Signing Secret); the value comes from the Slack app dashboard under Basic Information > App
// Credentials.
var errSlackSigningSecretNotConfigured = errors.New(
	"slack signing secret is not configured on the provider app; set it in the Ampersand dashboard")

// slackConfig is the per-provider subscribe-config bundle for Slack.
//
// Slack is not capable of managing subscriptions programmatically. It does carry webhook
// verification: the signing secret collected from the builder (stored in
// ProviderApp.metadata.providerParams) is threaded to the verifier per-request via paramsFn.
var slackConfig = ProviderConfig{
	Verification: VerificationConfig{
		paramsFn:          getSlackVerificationParams,
		verifierConnector: &slack.Connector{},
	},
}

// getSlackVerificationParams resolves the Slack signing secret from the provider app's metadata
// (ProviderApp.metadata.providerParams.webhookSigningSecret) into the verifier's VerificationParams.
func getSlackVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil {
		return nil, errInstallationNotFound
	}

	// ProviderParams is a *map on the wire type, so guard the pointer before indexing.
	secret := ""
	if req.ProviderApp != nil && req.ProviderApp.Metadata != nil &&
		req.ProviderApp.Metadata.ProviderParams != nil {
		secret = (*req.ProviderApp.Metadata.ProviderParams)[slack.ProviderParamWebhookSigningSecret]
	}

	if secret == "" {
		return nil, errSlackSigningSecretNotConfigured
	}

	return &common.VerificationParams{
		Param: &slack.VerificationParams{SigningSecret: secret},
	}, nil
}
