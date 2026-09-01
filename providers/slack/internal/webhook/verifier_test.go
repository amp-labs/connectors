package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

const testSigningKey = "3e81ee19b766670a1e6058fa895148ce"

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	eventMessage := testutils.DataFromFile(t, "event-for-verification.json")

	// Use a fresh timestamp to satisfy the 5-minute window check
	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	invalidTimestamp := strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10)

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
			},
			Server:   mockserver.Dummy(),
			Expected: true,
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

// TestVerifyWebhookMessageParamsSecret exercises the params-provided signing secret path used by
// the subscribe flow, where the verifier connector is a zero value and the secret arrives
// per-request via VerificationParams.
func TestVerifyWebhookMessageParamsSecret(t *testing.T) {
	t.Parallel()

	eventMessage := testutils.DataFromFile(t, "event-for-verification.json")

	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	validSlackSignature := computeSlackSignature(testSigningKey, validTimestamp, string(eventMessage))

	validRequest := &common.WebhookRequest{
		Headers: http.Header{
			"X-Slack-Signature":         []string{validSlackSignature},
			"X-Slack-Request-Timestamp": []string{validTimestamp},
		},
		Body: eventMessage,
	}

	tests := []testconn.TestCaseVerifyWebhookMessage{
		{
			Name: "Valid signature with params-provided secret on secretless verifier",
			Input: testconn.WebhookMessageVerificationParams{
				Request: validRequest,
				Params: &common.VerificationParams{
					Param: &VerificationParams{SigningSecret: testSigningKey},
				},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
		{
			Name: "Missing secret everywhere is rejected",
			Input: testconn.WebhookMessageVerificationParams{
				Request: validRequest,
				Params:  &common.VerificationParams{},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				ErrSigningSecretIsNotSet,
			},
		},
		{
			Name: "Wrong params type is rejected",
			Input: testconn.WebhookMessageVerificationParams{
				Request: validRequest,
				Params: &common.VerificationParams{
					Param: "not-a-verification-params-struct",
				},
			},
			Server:   mockserver.Dummy(),
			Expected: false,
			ExpectedErrs: []error{
				testutils.StringError("invalid verification params"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWebhookMessageVerifier, error) {
				return NewVerifier(""), nil
			})
		})
	}
}

// TestVerifyWebhookMessageParamsOverrideConstructor asserts the params secret wins over the
// constructor secret.
func TestVerifyWebhookMessageParamsOverrideConstructor(t *testing.T) {
	t.Parallel()

	eventMessage := testutils.DataFromFile(t, "event-for-verification.json")

	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	paramsSigningKey := "params-signing-key"
	signatureFromParamsKey := computeSlackSignature(paramsSigningKey, validTimestamp, string(eventMessage))

	tt := testconn.TestCaseVerifyWebhookMessage{
		Name: "Params secret overrides constructor secret",
		Input: testconn.WebhookMessageVerificationParams{
			Request: &common.WebhookRequest{
				Headers: http.Header{
					"X-Slack-Signature":         []string{signatureFromParamsKey},
					"X-Slack-Request-Timestamp": []string{validTimestamp},
				},
				Body: eventMessage,
			},
			Params: &common.VerificationParams{
				Param: &VerificationParams{SigningSecret: paramsSigningKey},
			},
		},
		Server:   mockserver.Dummy(),
		Expected: true,
	}

	tt.Run(t, func() (testconn.TestableWebhookMessageVerifier, error) {
		return constructTestVerifier()
	})
}

func constructTestVerifier() (*Verifier, error) {
	verifier := NewVerifier(testSigningKey)

	return verifier, nil
}

func computeSlackSignature(signingKey, timestamp, body string) string {
	sigBasestring := fmt.Sprintf("v0:%s:%s", timestamp, body)
	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(sigBasestring))
	return "v0=" + hex.EncodeToString(h.Sum(nil))
}
