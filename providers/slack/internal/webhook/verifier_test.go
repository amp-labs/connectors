package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const testSigningKey = "3e81ee19b766670a1e6058fa895148ce"

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	eventMessage := testutils.DataFromFile(t, "event-for-verification.json")

	// Use a fresh timestamp to satisfy the 5-minute window check
	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	invalidTimestamp := strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10)

	validSlackSignature := computeSlackSignature(testSigningKey, validTimestamp, string(eventMessage))
	invalidSlackSignature := "mismatching-signature-from-provider"

	tests := []testroutines.TestCaseVerifyWebhookMessage{
		{
			Name: "Missing signature header in input",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				common.ErrMissingHeader,
				testutils.StringError("header 'x-slack-signature'"),
			},
		},
		{
			Name: "Missing timestamp header in input",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature": []string{validSlackSignature},
					},
					Body: eventMessage,
				},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				common.ErrMissingHeader,
				testutils.StringError("header 'x-slack-request-timestamp'"),
			},
		},
		{
			Name: "Invalid signature",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{invalidSlackSignature},
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
		},
		{
			Name: "Invalid timestamp",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{validSlackSignature},
						"X-Slack-Request-Timestamp": []string{invalidTimestamp},
					},
					Body: eventMessage,
				},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				testutils.StringError("request timestamp is more than 5 minutes old"),
			},
		},
		{
			Name: "Valid signature",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{validSlackSignature},
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableWebhookMessageVerifier, error) {
				return constructTestVerifier(tt.Server)
			})
		})
	}
}

func constructTestVerifier(server *httptest.Server) (*Verifier, error) {
	transport, err := components.NewTransport(providers.ConnectWise, common.ConnectorParams{
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestMockServerBaseURL(server.URL)

	verifier := NewVerifier(transport.JSONHTTPClient(), transport.ProviderInfo(), testSigningKey)

	return verifier, nil
}

func computeSlackSignature(signingKey, timestamp, body string) string {
	sigBasestring := fmt.Sprintf("v0:%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(sigBasestring))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}
