package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

// computeTestSignature computes a Stripe signature for testing purposes.
func computeTestSignature(secret, timestamp string, body []byte) string {
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	secret := "whsec_test_secret"
	// Use a recent timestamp (current time) for valid tests
	recentTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	body := []byte(`{"id":"evt_test_123","type":"charge.succeeded"}`)

	validSignature := computeTestSignature(secret, recentTimestamp, body)
	validSignatureHeader := fmt.Sprintf("t=%s,v1=%s", recentTimestamp, validSignature)

	conn := &Connector{}

	tests := []struct {
		name          string
		request       *common.WebhookRequest
		params        *common.VerificationParams
		expectedValid bool
		expectedError error
	}{
		{
			name: "Valid signature",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{validSignatureHeader},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: true,
			expectedError: nil,
		},

		{
			name: "Invalid signature",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{fmt.Sprintf("t=%s,v1=%s", recentTimestamp, "invalid_signature")},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: false,
			expectedError: errInvalidSignature,
		},
		{
			name: "Missing signature header",
			request: &common.WebhookRequest{
				Headers: http.Header{},
				Body:    body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: false,
			expectedError: errMissingSignature,
		},
		{
			name:    "Nil request",
			request: nil,
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: false,
			expectedError: errMissingParams,
		},
		{
			name: "Empty secret",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{validSignatureHeader},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: "",
				},
			},
			expectedValid: false,
			expectedError: errMissingParams,
		},
		{
			name: "Wrong timestamp (signature mismatch)",
			request: &common.WebhookRequest{
				Headers: http.Header{
					// Use a valid recent timestamp but with a signature computed for a different timestamp
					"Stripe-Signature": []string{fmt.Sprintf("t=%s,v1=%s", recentTimestamp, computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-10), body))},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: false,
			expectedError: errInvalidSignature,
		},
		{
			name: "Timestamp too old (replay attack)",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{fmt.Sprintf("t=%d,v1=%s", time.Now().Unix()-600, computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-600), body))},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: false,
			expectedError: errTimestampTooOld,
		},
		{
			name: "Timestamp too far in the future",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{fmt.Sprintf("t=%d,v1=%s", time.Now().Unix()+600, computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()+600), body))},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: secret,
				},
			},
			expectedValid: false,
			expectedError: errTimestampTooFarInFuture,
		},
		{
			name: "Invalid tolerance (zero)",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{validSignatureHeader},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret:    secret,
					Tolerance: 0,
				},
			},
			expectedValid: true,
			expectedError: nil,
		},
		{
			name: "Invalid tolerance (negative)",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{validSignatureHeader},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret:    secret,
					Tolerance: -1 * time.Minute,
				},
			},
			expectedValid: false,
			expectedError: errInvalidTolerance,
		},
		{
			name: "Custom tolerance within limit",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{fmt.Sprintf("t=%d,v1=%s", time.Now().Unix()-120, computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-120), body))},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret:    secret,
					Tolerance: 5 * time.Minute, // 5 minutes tolerance
				},
			},
			expectedValid: true,
			expectedError: nil,
		},
		{
			name: "Custom tolerance exceeded",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{fmt.Sprintf("t=%d,v1=%s", time.Now().Unix()-120, computeTestSignature(secret, fmt.Sprintf("%d", time.Now().Unix()-120), body))},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret:    secret,
					Tolerance: 1 * time.Minute, // Only 1 minute tolerance, but timestamp is 2 minutes old
				},
			},
			expectedValid: false,
			expectedError: errTimestampTooOld,
		},
		{
			name: "Wrong secret",
			request: &common.WebhookRequest{
				Headers: http.Header{
					"Stripe-Signature": []string{validSignatureHeader},
				},
				Body: body,
			},
			params: &common.VerificationParams{
				Param: &VerificationParams{
					Secret: "wrong_secret",
				},
			},
			expectedValid: false,
			expectedError: errInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := conn.VerifyWebhookMessage(context.Background(), tt.request, tt.params)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "should return expected error")
				assert.Equal(t, valid, false, "should return false for invalid verification")
			} else {
				assert.NilError(t, err, "should not return error")
				assert.Equal(t, valid, tt.expectedValid, "verification result should match expected")
			}
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
			name:         "Parse signature header",
			header:       "t=1766887044,v1=08ddffb964639dd31625fa74a9fcb8e95daaef2220ebd8e493127cf2a06320f7,v0=2c1d2ba92e6f80203fbbd6b46b9b2386693bb0d4a1987432c4646e817583201d",
			expectedTS:   "1766887044",
			expectedSigs: []string{"08ddffb964639dd31625fa74a9fcb8e95daaef2220ebd8e493127cf2a06320f7", "2c1d2ba92e6f80203fbbd6b46b9b2386693bb0d4a1987432c4646e817583201d"},
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
