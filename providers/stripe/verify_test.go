package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"gotest.tools/v3/assert"
)

// computeTestSignature computes a Stripe signature for testing purposes.
func computeTestSignature(secret, timestamp string, body []byte) string {
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))

	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookMessage(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	secret := "whsec_test_secret"
	// Use a recent timestamp (current time) for valid tests
	recentTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	body := []byte(`{"id":"evt_test_123","type":"charge.succeeded"}`)

	validSignature := computeTestSignature(secret, recentTimestamp, body)
	validSignatureHeader := fmt.Sprintf("t=%s,v1=%s", recentTimestamp, validSignature)

	tests := []testconn.TestCaseVerifyWebhookMessage{
		{
			Name: "Valid signature",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{validSignatureHeader},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
		{
			Name: "Invalid signature",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{fmt.Sprintf("t=%s,v1=%s", recentTimestamp, "invalid_signature")},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errInvalidSignature},
		},
		{
			Name: "Missing signature header",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{},
					Body:    body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errMissingSignature},
		},
		{
			Name: "Nil request",
			Input: testconn.WebhookMessageVerificationParams{
				Request: nil,
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Empty secret",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{validSignatureHeader},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: "",
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Wrong timestamp (signature mismatch)",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						// Use a valid recent timestamp but with a signature computed for a different timestamp
						"Stripe-Signature": []string{fmt.Sprintf(
							"t=%s,v1=%s",
							recentTimestamp,
							computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-10), body),
						)},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errInvalidSignature},
		},
		{
			Name: "Timestamp too old (replay attack)",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{fmt.Sprintf(
							"t=%d,v1=%s",
							time.Now().Unix()-600,
							computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-600), body),
						)},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errTimestampTooOld},
		},
		{
			Name: "Timestamp too far in the future",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{fmt.Sprintf(
							"t=%d,v1=%s",
							time.Now().Unix()+600,
							computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()+600), body),
						)},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: secret,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errTimestampTooFarInFuture},
		},
		{
			Name: "Default tolerance applies when unset (zero)",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{validSignatureHeader},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret:    secret,
						Tolerance: 0,
					},
				},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
		{
			Name: "Invalid tolerance (negative)",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{validSignatureHeader},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret:    secret,
						Tolerance: -1 * time.Minute,
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errInvalidTolerance},
		},
		{
			Name: "Custom tolerance within limit",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{fmt.Sprintf(
							"t=%d,v1=%s",
							time.Now().Unix()-120,
							computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-120), body),
						)},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret:    secret,
						Tolerance: 5 * time.Minute, // 5 minutes tolerance
					},
				},
			},
			Server:   mockserver.Dummy(),
			Expected: true,
		},
		{
			Name: "Custom tolerance exceeded",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{fmt.Sprintf(
							"t=%d,v1=%s",
							time.Now().Unix()-120,
							computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-120), body),
						)},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret:    secret,
						Tolerance: 1 * time.Minute, // Only 1 minute tolerance, but timestamp is 2 minutes old
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errTimestampTooOld},
		},
		{
			Name: "Wrong secret",
			Input: testconn.WebhookMessageVerificationParams{
				Request: &common.WebhookRequest{
					Headers: http.Header{
						"Stripe-Signature": []string{validSignatureHeader},
					},
					Body: body,
				},
				Params: &common.VerificationParams{
					Param: &VerificationParams{
						Secret: "wrong_secret",
					},
				},
			},
			Server:       mockserver.Dummy(),
			Expected:     false,
			ExpectedErrs: []error{errInvalidSignature},
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWebhookMessageVerifier, error) {
				return &Connector{}, nil
			})
		})
	}
}

func TestParseStripeSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		header        string
		expectedTS    string
		expectedSigs  []string
		expectedError string
	}{
		{
			name: "Parse signature header",
			header: "t=1766887044,v1=08ddffb964639dd31625fa74a9fcb8e95daaef2220ebd8e493127cf2a06320f7," +
				"v0=2c1d2ba92e6f80203fbbd6b46b9b2386693bb0d4a1987432c4646e817583201d",
			expectedTS: "1766887044",
			expectedSigs: []string{
				"08ddffb964639dd31625fa74a9fcb8e95daaef2220ebd8e493127cf2a06320f7",
				"2c1d2ba92e6f80203fbbd6b46b9b2386693bb0d4a1987432c4646e817583201d",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, sigs, err := parseStripeSignature(tt.header)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError, "should return expected error")
			} else {
				assert.NilError(t, err, "should not return error")
				assert.Equal(t, ts, tt.expectedTS, "timestamp should match")
				assert.DeepEqual(t, sigs, tt.expectedSigs)
			}
		})
	}
}
