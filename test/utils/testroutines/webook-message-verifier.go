package testroutines

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

type (
	WebhookMessageVerificationType = TestCase[WebhookMessageVerificationParams, bool]
	// WebhookMessageVerification is a test suite useful for testing connectors.WebhookVerifierConnector interface.
	WebhookMessageVerification WebhookMessageVerificationType
)

type WebhookMessageVerificationParams struct {
	Request *common.WebhookRequest
	Params  *common.VerificationParams
}

// Run provides a procedure to test connectors.WebhookVerifierConnector
func (r WebhookMessageVerification) Run(t *testing.T, builder ConnectorBuilder[components.WebhookMessageVerifier]) {
	t.Helper()
	t.Cleanup(func() {
		WebhookMessageVerificationType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	output, err := conn.VerifyWebhookMessage(t.Context(), r.Input.Request, r.Input.Params)
	WebhookMessageVerificationType(r).Validate(t, err, output)
}
