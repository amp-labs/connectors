package stripe

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// Decomposed per-method interface assertions, so Subscribe, UpdateSubscription and
// DeleteSubscription are each verified independently (see pr-4-subscribe-update-delete.md).
var (
	_ testconn.TestableSubscriptionCreator = &Connector{} // Subscribe
	_ testconn.TestableSubscriptionUpdater = &Connector{} // UpdateSubscription
	_ testconn.TestableSubscriptionRemover = &Connector{} // DeleteSubscription
)

func TestSubscribe(t *testing.T) { // nolint:funlen
	t.Parallel()

	webhookEndpointResponse := testutils.DataFromFile(t, "subscribe/webhook-endpoint-response.json")

	tests := []testconn.TestCaseSubscribe{
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
					"accounts": {
						Events: []common.SubscriptionEventType{
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
					mockcond.Body(`
						enabled_events[]=account.updated&
						url=https://webhook.site/test`,
					),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointResponse),
			}.Server(),
			ExpectedErrs: nil,
			Comparator:   compareSubscriptionSuccess,
		},
		{
			Name: "Subscribe multiple objects",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"accounts": {
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
					"customers": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
							common.SubscriptionEventTypeDelete,
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
					mockcond.Body(`
						enabled_events[]=account.application.authorized&
						enabled_events[]=account.application.deauthorized&
						enabled_events[]=account.updated&
						enabled_events[]=balance.available&
						enabled_events[]=billing_portal.configuration.created&
						enabled_events[]=charge.dispute.funds_withdrawn&
						enabled_events[]=charge.succeeded&
						enabled_events[]=customer.created&
						enabled_events[]=customer.deleted&
						enabled_events[]=customer.updated&
						url=https://webhook.site/test`,
					),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointResponse),
			}.Server(),
			ExpectedErrs: nil,
			Comparator:   compareSubscriptionSuccess,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionCreator, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

// compareSubscriptionSuccess verifies the operation returned a successful subscription result.
func compareSubscriptionSuccess(
	_ string, actual, _ *common.SubscriptionResult,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	if actual == nil {
		return result.AddDiff("subscription result is nil")
	}

	result.Assert("Status", common.SubscriptionStatusSuccess, actual.Status)

	return result
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
				"accounts": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
				},
			},
			expected:    map[string]bool{"account.updated": true},
			expectedErr: nil,
			description: "Test building event set from single object with single event",
		},
		{
			name: "Single object with pass-through event",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"accounts": {
					PassThroughEvents: []string{"account.application.authorized"},
				},
			},
			expected:    map[string]bool{"account.application.authorized": true},
			expectedErr: nil,
			description: "Test building event set with pass-through event",
		},
		{
			name: "Object with no events is rejected",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"accounts": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
				},
				"refunds": {},
			},
			expected:    nil,
			expectedErr: errMissingParams,
			description: "An object resolving to zero events must error instead of silently reporting success",
		},
		{
			name: "Multiple objects, multiple events",
			subscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"accounts": {
					Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
					PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
				},
				"refunds": {
					PassThroughEvents: []string{"refund.created"},
				},
				"billing/meters": {
					PassThroughEvents: []string{"billing.meter.deactivated"},
				},
				"issuing/disputes": {
					PassThroughEvents: []string{"issuing_dispute.submitted", "issuing_dispute.funds_reinstated"},
				},
			},
			expected: map[string]bool{
				"account.application.authorized":   true,
				"account.application.deauthorized": true,
				"account.updated":                  true,
				"refund.created":                   true,
				"billing.meter.deactivated":        true,
				"issuing_dispute.submitted":        true,
				"issuing_dispute.funds_reinstated": true,
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
				"accounts": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
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
				"accounts": {
					Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
				},
				"tax_ids": {
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
				"accounts": {
					Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
					PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
				},
				"refunds": {
					PassThroughEvents: []string{"refund.created"},
				},
				"billing/meters": {
					PassThroughEvents: []string{"billing.meter.deactivated"},
				},
				"issuing/disputes": {
					PassThroughEvents: []string{"issuing_dispute.submitted", "issuing_dispute.funds_reinstated"},
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

func TestBuildSubscriptionResultDedup(t *testing.T) {
	t.Parallel()

	response := &WebhookResponse{
		ID:            "we_123",
		EnabledEvents: []string{"account.updated"},
	}

	// The same Stripe event expressed through both Events and PassThroughEvents
	// must be stored once, matching the deduplicated enabled_events sent to Stripe.
	subscriptionEvents := map[common.ObjectName]common.ObjectEvents{
		"accounts": {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
			PassThroughEvents: []string{"account.updated"},
		},
	}

	result, err := buildSubscriptionResult(response, subscriptionEvents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subResult, ok := result.Result.(*SubscriptionResult)
	if !ok {
		t.Fatalf("expected SubscriptionResult, got %T", result.Result)
	}

	enabledEvents := subResult.Subscriptions["accounts"].EnabledEvents
	if len(enabledEvents) != 1 || enabledEvents[0] != "account.updated" {
		t.Errorf("expected deduplicated [account.updated], got %v", enabledEvents)
	}
}
