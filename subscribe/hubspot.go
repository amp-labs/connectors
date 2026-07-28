package subscribe

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// hubspotConfig is the per-provider subscribe-config bundle for Hubspot. Hubspot is look-up-only
// (subscriptions configured at the provider-app level, so no subscribe/registration declarations),
// but it carries webhook-verification data: a verification-params builder, a verifier connector, a
// signature-verification bypass (public-app verification not yet implemented), and an event caster
// for its webhook payloads.
var hubspotConfig = ProviderConfig{
	Verification: VerificationConfig{
		paramsFn:          getHubspotVerificationParams,
		verifierConnector: &hubspot.Connector{},
		bypassed:          true,
		eventCaster:       CastSubscriptionEvents[hubspot.SubscriptionEvent],
	},
}

func getHubspotVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil {
		return nil, errInstallationNotFound
	}

	return &common.VerificationParams{
		Param: &hubspot.HubspotVerificationParams{
			ClientSecret: req.ProviderAppClientSecret,
		},
	}, nil
}
