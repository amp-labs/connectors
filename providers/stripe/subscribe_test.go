package stripe

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
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
