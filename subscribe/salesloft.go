package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// salesloftConfig is the per-provider subscribe-config bundle for Salesloft. Salesloft subscribes
// via API and builds a subscribe-time request payload (webhook endpoint + secret), so a
// buildRequestFn is declared. Salesloft is also one of the providers that requires
// WatchFieldsAuto="all" for subscribe update events.
var salesloftConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn:          getSalesloftRequest,
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		paramsFn:          getSalesloftVerificationParams,
		verifierConnector: &salesloft.Connector{},
	},
}

func getSalesloftVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil || req.Installation == nil {
		return nil, errInstallationNotFound
	}

	return &common.VerificationParams{
		Param: &salesloft.SalesloftVerificationParams{
			Secret: req.Installation.Id,
		},
	}, nil
}

func getSalesloftRequest(
	_ context.Context,
	_ deps.Dependencies,
	inst *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &salesloft.SubscriptionRequest{
		UniqueRef:       "amp_" + inst.Id,
		WebhookEndPoint: webhookURL,
		Secret:          inst.Id,
	}, nil
}
