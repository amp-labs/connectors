package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/outreach"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// outreachConfig is the per-provider subscribe-config bundle for Outreach. Outreach subscribes
// via API and builds a subscribe-time request payload (webhook endpoint + secret), so a
// buildRequestFn is declared.
var outreachConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getOutreachRequest,
	},
	Verification: VerificationConfig{
		paramsFn:          getOutreachVerificationParams,
		verifierConnector: &outreach.Connector{},
	},
}

func getOutreachVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil || req.Installation == nil {
		return nil, errInstallationNotFound
	}

	return &common.VerificationParams{
		Param: &outreach.OutreachVerificationParams{
			Secret: req.Installation.Id,
		},
	}, nil
}

func getOutreachRequest(
	_ context.Context,
	_ deps.Dependencies,
	inst *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &outreach.SubscriptionRequest{
		UniqueRef:       "amp_" + inst.Id,
		WebhookEndPoint: webhookURL,
		Secret:          inst.Id,
	}, nil
}
