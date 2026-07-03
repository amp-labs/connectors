package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
)

type (
	webhookMessageVerificationType = TestCase[WebhookMessageVerificationParams, bool]
	// TestCaseVerifyWebhookMessage is a test suite useful for testing connectors.WebhookVerifierConnector interface.
	TestCaseVerifyWebhookMessage webhookMessageVerificationType
)

type WebhookMessageVerificationParams struct {
	Request *common.WebhookRequest
	Params  *common.VerificationParams
}

// Run provides a procedure to test connectors.WebhookVerifierConnector
func (r TestCaseVerifyWebhookMessage) Run(t *testing.T,
	builder ConnectorBuilder[TestableWebhookMessageVerifier],
) {
	t.Helper()
	t.Cleanup(func() {
		webhookMessageVerificationType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	input := webhookMessageVerificationType(r).PrepareInput()
	output, err := conn.VerifyWebhookMessage(t.Context(), r.Input.Request, input.Params)
	webhookMessageVerificationType(r).Validate(t, err, output)
}
