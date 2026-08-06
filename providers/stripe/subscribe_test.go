package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

func TestSubscribe(t *testing.T) {
	t.Parallel()

	webhookEndpointResponse := testutils.DataFromFile(t, "subscribe/webhook-endpoint-response.json")

	tests := []testroutines.TestCase[common.SubscribeParams, *common.SubscriptionResult]{

		{
			Name: "Empty events",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{},
				Request: &SubscriptionRequest{
					WebhookEndPoint: "https://webhook.site/test",
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Subscribe single object",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
				Request: &SubscriptionRequest{
					WebhookEndPoint: "https://webhook.site/test",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/webhook_endpoints"),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointResponse),
			}.Server(),
			ExpectedErrs: nil,
			Comparator: func(_ string, actual, expected *common.SubscriptionResult) bool {
				return actual != nil && actual.Status == common.SubscriptionStatusSuccess
			},
		},
		{
			Name: "Subscribe multiple objects",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
						PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
					},
					"balance": {
						PassThroughEvents: []string{"balance.available"},
					},
					"billing_portal": {
						PassThroughEvents: []string{"billing_portal.configuration.created"},
					},
					"charge": {
						PassThroughEvents: []string{"charge.dispute.funds_withdrawn", "charge.succeeded"},
					},
				},
				Request: &SubscriptionRequest{
					WebhookEndPoint: "https://webhook.site/test",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/webhook_endpoints"),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointResponse),
			}.Server(),
			ExpectedErrs: nil,
			Comparator: func(_ string, actual, expected *common.SubscriptionResult) bool {
				return actual != nil && actual.Status == common.SubscriptionStatusSuccess
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.Subscribe(t.Context(), tt.Input)
			tt.Validate(t, err, result)
		})
	}
}

func TestBuildRequestedEventSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		subscriptionEvents map[common.ObjectName]common.ObjectEvents
		expected           map[string]bool
		expectedErr        error
		description        string
	}{
		{
			name:               "Empty events",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{},
			expected:           map[string]bool{},
			expectedErr:        nil,
			description:        "Test building event set from empty subscription events",
		},
		{
			name: "Single object, single event",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
				},
			},
			expected:    map[string]bool{"account.created": true},
			expectedErr: nil,
			description: "Test building event set from single object with single event",
		},
		{
			name: "Single object with pass-through event",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					PassThroughEvents: []string{"account.application.authorized"},
				},
			},
			expected:    map[string]bool{"account.application.authorized": true},
			expectedErr: nil,
			description: "Test building event set with pass-through event",
		},
		{
			name: "Multiple objects, multiple events",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
					PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
				},
				"balance": {
					PassThroughEvents: []string{"balance.available"},
				},
				"billing_portal": {
					PassThroughEvents: []string{"billing_portal.configuration.created"},
				},
				"charge": {
					PassThroughEvents: []string{"charge.dispute.funds_withdrawn", "charge.succeeded"},
				},
			},
			expected: map[string]bool{
				"account.application.authorized":       true,
				"account.application.deauthorized":     true,
				"account.updated":                      true,
				"balance.available":                    true,
				"billing_portal.configuration.created": true,
				"charge.dispute.funds_withdrawn":       true,
				"charge.succeeded":                     true,
			},
			expectedErr: nil,
			description: "Test building event set with all sample events using pass-through",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildRequestedEventSet(tt.subscriptionEvents)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("expected %d events, got %d", len(tt.expected), len(result))
				}
				for event, expected := range tt.expected {
					if result[event] != expected {
						t.Errorf("expected event %s to be %v, got %v", event, expected, result[event])
					}
				}
			}
		})
	}
}

func TestBuildSubscriptionResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		response           *WebhookResponse
		subscriptionEvents map[common.ObjectName]common.ObjectEvents
		expectedCount      int
		description        string
	}{
		{
			name: "Single object",
			response: &WebhookResponse{
				ID:            "we_123",
				EnabledEvents: []string{"account.created"},
			},
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
				},
			},
			expectedCount: 1,
			description:   "Test building subscription result for single object",
		},
		{
			name: "Multiple objects",
			response: &WebhookResponse{
				ID:            "we_123",
				EnabledEvents: []string{"account.created", "charge.created"},
			},
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
				},
				"charge": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
				},
			},
			expectedCount: 2,
			description:   "Test building subscription result for multiple objects",
		},
		{
			name: "Multiple objects with pass-through events",
			response: &WebhookResponse{
				ID: "we_123",
				EnabledEvents: []string{
					"account.application.authorized",
					"account.application.deauthorized",
					"account.updated",
					"balance.available",
					"billing_portal.configuration.created",
					"charge.dispute.funds_withdrawn",
					"charge.succeeded",
				},
			},
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
					PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
				},
				"balance": {
					PassThroughEvents: []string{"balance.available"},
				},
				"billing_portal": {
					PassThroughEvents: []string{"billing_portal.configuration.created"},
				},
				"charge": {
					PassThroughEvents: []string{"charge.dispute.funds_withdrawn", "charge.succeeded"},
				},
			},
			expectedCount: 4,
			description:   "Test building subscription result with pass-through events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildSubscriptionResult(tt.response, tt.subscriptionEvents)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			if result.Status != common.SubscriptionStatusSuccess {
				t.Errorf("expected status success, got %s", result.Status)
			}
			subResult, ok := result.Result.(*SubscriptionResult)
			if !ok {
				t.Fatalf("expected SubscriptionResult, got %T", result.Result)
			}
			if len(subResult.Subscriptions) != tt.expectedCount {
				t.Errorf("expected %d subscriptions, got %d", tt.expectedCount, len(subResult.Subscriptions))
			}
			// Verify all subscriptions have composite IDs with format endpointID:objectName
			for obj, endpoint := range subResult.Subscriptions {
				expectedID := fmt.Sprintf("%s:%s", tt.response.ID, string(obj))
				if endpoint.ID != expectedID {
					t.Errorf("expected endpoint ID %s for object %s, got %s", expectedID, obj, endpoint.ID)
				}
			}
		})
	}
}

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
