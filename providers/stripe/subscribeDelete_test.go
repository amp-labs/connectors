package stripe

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDeleteSubscription(t *testing.T) {
	t.Parallel()

	responseWithAccountEventsOnly := testutils.DataFromFile(t, "subscribe/webhook-endpoint-account-events-only.json")
	responseWithMultipleEvents := testutils.DataFromFile(t, "subscribe/webhook-endpoint-multiple-events.json")
	responseAfterPartialDelete := testutils.DataFromFile(t, "subscribe/webhook-endpoint-after-partial-delete.json")

	tests := []testroutines.TestCase[common.SubscriptionResult, error]{
		{
			Name:         "Nil result",
			Input:        common.SubscriptionResult{Result: nil},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Empty subscriptions",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Delete all events - deletes endpoint",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"account": {
							ID:            "we_123:account",
							EnabledEvents: []string{"account.application.authorized", "account.updated"},
						},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/webhook_endpoints/we_123"),
						},
						Then: mockserver.Response(http.StatusOK, responseWithAccountEventsOnly),
					},
					{
						If: mockcond.And{
							mockcond.MethodDELETE(),
							mockcond.Path("/v1/webhook_endpoints/we_123"),
						},
						Then: mockserver.Response(http.StatusOK),
					},
				},
			}.Server(),
			ExpectedErrs: nil,
		},
		{
			Name: "Delete partial events - updates endpoint",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"account": {
							ID:            "we_123:account",
							EnabledEvents: []string{"account.application.authorized", "account.updated"},
						},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/webhook_endpoints/we_123"),
						},
						Then: mockserver.Response(http.StatusOK, responseWithMultipleEvents),
					},
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v1/webhook_endpoints/we_123"),
						},
						Then: mockserver.Response(http.StatusOK, responseAfterPartialDelete),
					},
				},
			}.Server(),
			ExpectedErrs: nil,
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

			err = conn.DeleteSubscription(t.Context(), tt.Input)
			tt.Validate(t, err, nil)
		})
	}
}

func TestValidateSubscriptionResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       common.SubscriptionResult
		expectedErr error
		description string
	}{
		{
			name:        "Nil result",
			input:       common.SubscriptionResult{Result: nil},
			expectedErr: errMissingParams,
			description: "Test validation with nil result field",
		},
		{
			name: "Invalid result type",
			input: common.SubscriptionResult{
				Result: "invalid type",
			},
			expectedErr: errInvalidRequestType,
			description: "Test validation with invalid result type",
		},
		{
			name: "Empty subscriptions",
			input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{},
				},
			},
			expectedErr: errMissingParams,
			description: "Test validation with empty subscriptions map",
		},
		{
			name: "Valid result",
			input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"account": {
							ID:            "we_123:account",
							EnabledEvents: []string{"account.updated"},
						},
					},
				},
			},
			expectedErr: nil,
			description: "Test validation with valid subscription result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateSubscriptionResult(tt.input)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExtractEndpointInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *SubscriptionResult
		expectedID  string
		expectedErr error
		description string
	}{
		{
			name: "Single object with composite ID",
			input: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.updated"},
					},
				},
			},
			expectedID:  "we_123",
			expectedErr: nil,
			description: "Test extracting endpoint ID from single composite ID",
		},
		{
			name: "Multiple objects with same endpoint",
			input: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.updated"},
					},
					"charge": {
						ID:            "we_123:charge",
						EnabledEvents: []string{"charge.succeeded"},
					},
				},
			},
			expectedID:  "we_123",
			expectedErr: nil,
			description: "Test extracting endpoint ID from multiple objects with same endpoint",
		},
		{
			name: "Backward compatible - no colon in ID",
			input: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123",
						EnabledEvents: []string{"account.updated"},
					},
				},
			},
			expectedID:  "we_123",
			expectedErr: nil,
			description: "Test extracting endpoint ID from non-composite ID (backward compatibility)",
		},
		{
			name: "Different endpoint IDs",
			input: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.updated"},
					},
					"charge": {
						ID:            "we_456:charge",
						EnabledEvents: []string{"charge.succeeded"},
					},
				},
			},
			expectedID:  "",
			expectedErr: errInvalidRequestType,
			description: "Test error when subscriptions have different endpoint IDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractEndpointInfo(tt.input)
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
				if result.ID != tt.expectedID {
					t.Errorf("expected ID %s, got %s", tt.expectedID, result.ID)
				}
				if len(result.ObjectsToDelete) != len(tt.input.Subscriptions) {
					t.Errorf("expected %d objects to delete, got %d", len(tt.input.Subscriptions), len(result.ObjectsToDelete))
				}
			}
		})
	}
}

func TestCollectEventsToRemove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		subscriptionData *SubscriptionResult
		objectsToDelete  map[common.ObjectName]bool
		expectedEvents   map[string]bool
		description      string
	}{
		{
			name: "Single object with multiple events",
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.application.authorized", "account.updated"},
					},
				},
			},
			objectsToDelete: map[common.ObjectName]bool{"account": true},
			expectedEvents: map[string]bool{
				"account.application.authorized": true,
				"account.updated":                true,
			},
			description: "Test collecting events from single object",
		},
		{
			name: "Multiple objects with different events",
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.application.authorized", "account.updated"},
					},
					"charge": {
						ID:            "we_123:charge",
						EnabledEvents: []string{"charge.dispute.funds_withdrawn", "charge.succeeded"},
					},
				},
			},
			objectsToDelete: map[common.ObjectName]bool{"account": true, "charge": true},
			expectedEvents: map[string]bool{
				"account.application.authorized": true,
				"account.updated":                true,
				"charge.dispute.funds_withdrawn": true,
				"charge.succeeded":               true,
			},
			description: "Test collecting events from multiple objects",
		},
		{
			name: "Multiple objects with overlapping events",
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.updated"},
					},
					"balance": {
						ID:            "we_123:balance",
						EnabledEvents: []string{"balance.available"},
					},
				},
			},
			objectsToDelete: map[common.ObjectName]bool{"account": true, "balance": true},
			expectedEvents: map[string]bool{
				"account.updated":   true,
				"balance.available": true,
			},
			description: "Test collecting events from multiple objects with different events",
		},
		{
			name: "All sample events",
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID: "we_123:account",
						EnabledEvents: []string{
							"account.application.authorized",
							"account.application.deauthorized",
							"account.updated",
						},
					},
					"balance": {
						ID:            "we_123:balance",
						EnabledEvents: []string{"balance.available"},
					},
					"billing_portal": {
						ID:            "we_123:billing_portal",
						EnabledEvents: []string{"billing_portal.configuration.created"},
					},
					"charge": {
						ID: "we_123:charge",
						EnabledEvents: []string{
							"charge.dispute.funds_withdrawn",
							"charge.succeeded",
						},
					},
				},
			},
			objectsToDelete: map[common.ObjectName]bool{
				"account":        true,
				"balance":        true,
				"billing_portal": true,
				"charge":         true,
			},
			expectedEvents: map[string]bool{
				"account.application.authorized":       true,
				"account.application.deauthorized":     true,
				"account.updated":                      true,
				"balance.available":                    true,
				"billing_portal.configuration.created": true,
				"charge.dispute.funds_withdrawn":       true,
				"charge.succeeded":                     true,
			},
			description: "Test collecting all sample events from multiple objects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectEventsToRemove(tt.subscriptionData, tt.objectsToDelete)
			if len(result) != len(tt.expectedEvents) {
				t.Errorf("expected %d events, got %d", len(tt.expectedEvents), len(result))
			}
			for event := range tt.expectedEvents {
				if !result[event] {
					t.Errorf("expected event %s to be in result", event)
				}
			}
		})
	}
}

func TestFilterEventsToKeep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		currentEvents  []string
		eventsToRemove map[string]bool
		expectedEvents []string
		description    string
	}{
		{
			name:          "Remove all events",
			currentEvents: []string{"account.updated", "charge.succeeded"},
			eventsToRemove: map[string]bool{
				"account.updated":  true,
				"charge.succeeded": true,
			},
			expectedEvents: []string{},
			description:    "Test filtering removes all events",
		},
		{
			name:          "Remove partial events",
			currentEvents: []string{"account.updated", "balance.available", "charge.succeeded"},
			eventsToRemove: map[string]bool{
				"account.updated": true,
			},
			expectedEvents: []string{"balance.available", "charge.succeeded"},
			description:    "Test filtering removes only specified events",
		},
		{
			name:           "Keep all events when none to remove",
			currentEvents:  []string{"account.updated", "charge.succeeded"},
			eventsToRemove: map[string]bool{},
			expectedEvents: []string{"account.updated", "charge.succeeded"},
			description:    "Test filtering keeps all events when removal set is empty",
		},
		{
			name: "Remove multiple events with all sample events",
			currentEvents: []string{
				"account.application.authorized",
				"account.application.deauthorized",
				"account.updated",
				"balance.available",
				"billing_portal.configuration.created",
				"charge.dispute.funds_withdrawn",
				"charge.succeeded",
			},
			eventsToRemove: map[string]bool{
				"account.application.authorized": true,
				"account.updated":                true,
				"charge.succeeded":               true,
			},
			expectedEvents: []string{
				"account.application.deauthorized",
				"balance.available",
				"billing_portal.configuration.created",
				"charge.dispute.funds_withdrawn",
			},
			description: "Test filtering with all sample events removes specified subset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEventsToKeep(tt.currentEvents, tt.eventsToRemove)
			if len(result) != len(tt.expectedEvents) {
				t.Errorf("expected %d events, got %d", len(tt.expectedEvents), len(result))
			}
			resultMap := make(map[string]bool)
			for _, event := range result {
				resultMap[event] = true
			}
			for _, expectedEvent := range tt.expectedEvents {
				if !resultMap[expectedEvent] {
					t.Errorf("expected event %s to be in result", expectedEvent)
				}
			}
		})
	}
}

func TestGetWebhookURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		currentEndpoint  *WebhookResponse
		subscriptionData *SubscriptionResult
		expectedURL      string
		description      string
	}{
		{
			name: "URL from current endpoint",
			currentEndpoint: &WebhookResponse{
				URL: "https://webhook.site/test",
			},
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{},
			},
			expectedURL: "https://webhook.site/test",
			description: "Test extracting URL from current endpoint",
		},
		{
			name:            "URL from subscription data fallback",
			currentEndpoint: &WebhookResponse{URL: ""},
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:  "we_123:account",
						URL: "https://webhook.site/fallback",
					},
				},
			},
			expectedURL: "https://webhook.site/fallback",
			description: "Test extracting URL from subscription data when current endpoint has no URL",
		},
		{
			name:            "Empty URL when both are empty",
			currentEndpoint: &WebhookResponse{URL: ""},
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:  "we_123:account",
						URL: "",
					},
				},
			},
			expectedURL: "",
			description: "Test returning empty string when no URL available",
		},
		{
			name: "Prefer current endpoint URL over subscription data",
			currentEndpoint: &WebhookResponse{
				URL: "https://webhook.site/current",
			},
			subscriptionData: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:  "we_123:account",
						URL: "https://webhook.site/subscription",
					},
				},
			},
			expectedURL: "https://webhook.site/current",
			description: "Test current endpoint URL takes precedence over subscription data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWebhookURL(tt.currentEndpoint, tt.subscriptionData)
			if result != tt.expectedURL {
				t.Errorf("expected URL %s, got %s", tt.expectedURL, result)
			}
		})
	}
}
