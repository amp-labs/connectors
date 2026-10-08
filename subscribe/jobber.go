package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/jobber"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// jobberConfig is the per-provider subscribe-config bundle for Jobber. Jobber subscribes via API
// (one webhook endpoint per resolved topic) and builds a subscribe-time request payload carrying
// the webhook URL, so a buildRequestFn is declared. Jobber signs each webhook with
// Base64(HMAC-SHA256(app OAuth client secret, raw body)), so verification is real (not bypassed):
// a params builder supplies the client secret and the connector performs the check. An event caster
// unwraps its webhook payloads into typed SubscriptionEvents.
//
// Jobber requires WatchFieldsAuto="all" for subscribe update events: its webhook payloads carry only
// identifiers (no changed-field list), so there is no way to tell which fields changed. We therefore
// treat every update event as watching all fields and fetch the full record back via the API.
var jobberConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn:          getJobberRequest,
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		paramsFn:          getJobberVerificationParams,
		verifierConnector: &jobber.Connector{},
		eventCaster:       CastSubscriptionEvents[jobber.SubscriptionEvent],
	},
}

func getJobberRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &jobber.SubscriptionRequest{
		WebhookURL: webhookURL,
	}, nil
}

// getJobberVerificationParams supplies the secret used to verify incoming Jobber webhooks.
//
// Jobber signs each webhook request with an HMAC-SHA256 digest of the raw request body, keyed by
// the app's OAuth client secret, and sends it Base64-encoded in the X-Jobber-Hmac-SHA256 header.
// The connector's VerifyWebhookMessage recomputes that digest and compares it constant-time, so the
// only per-app input it needs is the client secret — which is the same OAuth client secret the app
// authenticates with, carried on deps.VerificationRequest.ProviderAppClientSecret (the openapi wire
// ProviderApp deliberately excludes secrets).
//
// Reference: https://developer.getjobber.com/docs/using_jobbers_api/setting_up_webhooks/
func getJobberVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil {
		return nil, errInstallationNotFound
	}

	return &common.VerificationParams{
		Param: &jobber.JobberVerificationParams{
			Secret: req.ProviderAppClientSecret,
		},
	}, nil
}
