package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	eventMessage := testutils.DataFromFile(t, "event-for-verification.json")

	// Use a fresh timestamp to satisfy the 5-minute window check
	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	invalidTimestamp := strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10)

	const testSigningKey = "3e81ee19b766670a1e6058fa895148ce"
	validSlackSignature := computeSlackSignature(testSigningKey, validTimestamp, string(eventMessage))
	invalidSlackSignature := "mismatching-signature-from-provider"

	tests := []testconn.TestCaseVerifyWebhookMessage{
		{
			Name: "Missing signature header in input",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
				Params: &common.VerificationParams{Param: &VerificationParams{SigningSecret: testSigningKey}},
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
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature": []string{validSlackSignature},
					},
					Body: eventMessage,
				},
				Params: &common.VerificationParams{Param: &VerificationParams{SigningSecret: testSigningKey}},
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
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{invalidSlackSignature},
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
				Params: &common.VerificationParams{Param: &VerificationParams{SigningSecret: testSigningKey}},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
		},
		{
			Name: "Invalid timestamp",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{validSlackSignature},
						"X-Slack-Request-Timestamp": []string{invalidTimestamp},
					},
					Body: eventMessage,
				},
				Params: &common.VerificationParams{Param: &VerificationParams{SigningSecret: testSigningKey}},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				testutils.StringError("request timestamp is more than 5 minutes old"),
			},
		},
		{
			Name: "Valid signature",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{validSlackSignature},
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
				Params: &common.VerificationParams{Param: &VerificationParams{SigningSecret: testSigningKey}},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
		{
			Name: "Verification param without the signing key",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"X-Slack-Signature":         []string{validSlackSignature},
						"X-Slack-Request-Timestamp": []string{validTimestamp},
					},
					Body: eventMessage,
				},
				Params: &common.VerificationParams{Param: &VerificationParams{SigningSecret: ""}},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				common.ErrMissingProviderParam,
				testutils.StringError("SigningSecret is empty"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWebhookMessageVerifier, error) {
				return constructTestVerifier()
			})
		})
	}
}

func constructTestVerifier() (*Verifier, error) {
	verifier := NewVerifier()

	return &verifier, nil
}

// TestVerifyWebhookMessageNilVerifier pins the backstop for an unconstructed verifier: a nil
// *Verifier (e.g. the nil embed of a zero-value slack.Connector) must return a clear error from
// the promoted call, never panic in the method-promotion wrapper.
func TestVerifyWebhookMessageNilVerifier(t *testing.T) {
	t.Parallel()

	var v *Verifier

	ok, err := v.VerifyWebhookMessage(context.Background(), &common.WebhookRequest{}, nil)
	if ok {
		t.Error("VerifyWebhookMessage() = true, want false on nil verifier")
	}

	if !errors.Is(err, errVerifierNotInitialized) {
		t.Errorf("VerifyWebhookMessage() error = %v, want errVerifierNotInitialized", err)
	}
}

func computeSlackSignature(signingKey, timestamp, body string) string {
	sigBasestring := fmt.Sprintf("v0:%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(sigBasestring))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}
