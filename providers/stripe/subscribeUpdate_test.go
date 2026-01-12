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

// UpdateSubscriptionInput contains both params and previousResult needed for UpdateSubscription
type UpdateSubscriptionInput struct {
	Params         common.SubscribeParams
	PreviousResult *common.SubscriptionResult
}

func TestUpdateSubscription(t *testing.T) {
	t.Parallel()

	webhookEndpointUpdatedResponse := testutils.DataFromFile(t, "subscribe/webhook-endpoint-updated-response.json")

	tests := []testroutines.TestCase[UpdateSubscriptionInput, *common.SubscriptionResult]{
		{
			Name: "Missing previous result",
			Input: UpdateSubscriptionInput{
				Params: common.SubscribeParams{
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"account": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
						},
					},
				},
				PreviousResult: nil,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Nil result field",
			Input: UpdateSubscriptionInput{
				Params: common.SubscribeParams{
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"account": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
						},
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result: nil,
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Invalid previous result type",
			Input: UpdateSubscriptionInput{
				Params: common.SubscribeParams{
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"account": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
						},
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result: "invalid type",
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errInvalidRequestType},
		},
		{
			Name: "Update single object events",
			Input: UpdateSubscriptionInput{
				Params: common.SubscribeParams{
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
				PreviousResult: &common.SubscriptionResult{
					Result: &SubscriptionResult{
						Subscriptions: map[common.ObjectName]WebhookResponse{
							"account": {
								ID:            "we_123:account",
								EnabledEvents: []string{"account.created"},
							},
						},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/webhook_endpoints/we_123"),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointUpdatedResponse),
			}.Server(),
			ExpectedErrs: nil,
			Comparator: func(_ string, actual, expected *common.SubscriptionResult) bool {
				return actual != nil && actual.Status == common.SubscriptionStatusSuccess
			},
		},
		{
			Name: "Update multiple objects with different events",
			Input: UpdateSubscriptionInput{
				Params: common.SubscribeParams{
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"account": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
								common.SubscriptionEventTypeUpdate,
							},
						},
						"charge": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
							},
						},
					},
					Request: &SubscriptionRequest{
						WebhookEndPoint: "https://webhook.site/test",
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result: &SubscriptionResult{
						Subscriptions: map[common.ObjectName]WebhookResponse{
							"balance": {
								ID:            "we_123:balance",
								EnabledEvents: []string{"balance.created"},
							},
						},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/webhook_endpoints/we_123"),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointUpdatedResponse),
			}.Server(),
			ExpectedErrs: nil,
			Comparator: func(_ string, actual, expected *common.SubscriptionResult) bool {
				return actual != nil && actual.Status == common.SubscriptionStatusSuccess
			},
		},
		{
			Name: "Update keeps existing objects not in params",
			Input: UpdateSubscriptionInput{
				Params: common.SubscribeParams{
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"charge": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
							},
						},
					},
					Request: &SubscriptionRequest{
						WebhookEndPoint: "https://webhook.site/test",
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result: &SubscriptionResult{
						Subscriptions: map[common.ObjectName]WebhookResponse{
							"account": {
								ID:            "we_123:account",
								EnabledEvents: []string{"account.created", "account.updated"},
							},
							"balance": {
								ID:            "we_123:balance",
								EnabledEvents: []string{"balance.created"},
							},
						},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/webhook_endpoints/we_123"),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointUpdatedResponse),
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

			result, err := conn.UpdateSubscription(t.Context(), tt.Input.Params, tt.Input.PreviousResult)
			tt.Validate(t, err, result)
		})
	}
}

func TestValidatePreviousResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		previousResult *common.SubscriptionResult
		expectedErr    error
		description    string
	}{
		{
			name:           "Nil previous result",
			previousResult: nil,
			expectedErr:    errMissingParams,
			description:    "Test validation with nil previous result",
		},
		{
			name: "Nil result field",
			previousResult: &common.SubscriptionResult{
				Result: nil,
			},
			expectedErr: errMissingParams,
			description: "Test validation with nil result field",
		},
		{
			name: "Invalid result type",
			previousResult: &common.SubscriptionResult{
				Result: "invalid",
			},
			expectedErr: errInvalidRequestType,
			description: "Test validation with invalid result type",
		},
		{
			name: "Valid result",
			previousResult: &common.SubscriptionResult{
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
			_, err := validatePreviousResult(tt.previousResult)
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

func TestGetExistingEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		subscriptions map[common.ObjectName]WebhookResponse
		expectedErr   error
		expectedID    string
		description   string
	}{
		{
			name:          "Empty subscriptions",
			subscriptions: map[common.ObjectName]WebhookResponse{},
			expectedErr:   errMissingParams,
			description:   "Test extracting endpoint from empty subscriptions",
		},
		{
			name: "Single subscription with composite ID",
			subscriptions: map[common.ObjectName]WebhookResponse{
				"account": {
					ID:            "we_123:account",
					EnabledEvents: []string{"account.updated"},
				},
			},
			expectedErr: nil,
			expectedID:  "we_123",
			description: "Test extracting endpoint ID from single composite ID",
		},
		{
			name: "Multiple subscriptions with same endpoint",
			subscriptions: map[common.ObjectName]WebhookResponse{
				"account": {
					ID:            "we_123:account",
					EnabledEvents: []string{"account.updated"},
				},
				"charge": {
					ID:            "we_123:charge",
					EnabledEvents: []string{"charge.created"},
				},
			},
			expectedErr: nil,
			expectedID:  "we_123",
			description: "Test extracting endpoint ID from multiple subscriptions with same endpoint",
		},
		{
			name: "Backward compatible - no colon in ID",
			subscriptions: map[common.ObjectName]WebhookResponse{
				"account": {
					ID:            "we_123",
					EnabledEvents: []string{"account.updated"},
				},
			},
			expectedErr: nil,
			expectedID:  "we_123",
			description: "Test extracting endpoint ID from non-composite ID (backward compatibility)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getExistingEndpoint(tt.subscriptions)
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
			}
		})
	}
}

func TestBuildMergedEventNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		prevState      *SubscriptionResult
		params         common.SubscribeParams
		expectedEvents []string
		expectedErr    error
		description    string
	}{
		{
			name: "Keep existing events and add new",
			prevState: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"balance": {
						ID:            "we_123:balance",
						EnabledEvents: []string{"balance.created"},
					},
				},
			},
			params: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
			},
			expectedEvents: []string{"balance.created", "account.created", "account.updated"},
			expectedErr:    nil,
			description:    "Test merging keeps existing events and adds new object events",
		},
		{
			name: "Update existing object events",
			prevState: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.created"},
					},
					"balance": {
						ID:            "we_123:balance",
						EnabledEvents: []string{"balance.created"},
					},
				},
			},
			params: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
			},
			expectedEvents: []string{"balance.created", "account.created", "account.updated"},
			expectedErr:    nil,
			description:    "Test merging updates existing object events while keeping others",
		},
		{
			name: "All sample events",
			prevState: &SubscriptionResult{
				Subscriptions: map[common.ObjectName]WebhookResponse{
					"account": {
						ID:            "we_123:account",
						EnabledEvents: []string{"account.application.authorized", "account.application.deauthorized", "account.updated"},
					},
				},
			},
			params: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
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
			},
			expectedEvents: []string{
				"account.application.authorized",
				"account.application.deauthorized",
				"account.updated",
				"balance.available",
				"billing_portal.configuration.created",
				"charge.dispute.funds_withdrawn",
				"charge.succeeded",
			},
			expectedErr: nil,
			description: "Test merging with all sample event types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildMergedEventNames(tt.prevState, tt.params)
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
			}
		})
	}
}

func TestBuildMergedSubscriptionEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		previousResult  *common.SubscriptionResult
		params          common.SubscribeParams
		expectedObjects []common.ObjectName
		description     string
	}{
		{
			name: "Keep existing objects and add new",
			previousResult: &common.SubscriptionResult{
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"balance": {
						Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
					},
				},
			},
			params: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
					},
				},
			},
			expectedObjects: []common.ObjectName{"balance", "account"},
			description:     "Test merging keeps existing objects and adds new",
		},
		{
			name: "Update existing object",
			previousResult: &common.SubscriptionResult{
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
					},
				},
			},
			params: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"account": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
			},
			expectedObjects: []common.ObjectName{"account"},
			description:     "Test merging updates existing object events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildMergedSubscriptionEvents(tt.previousResult, tt.params)
			if len(result) != len(tt.expectedObjects) {
				t.Errorf("expected %d objects, got %d", len(tt.expectedObjects), len(result))
			}
			for _, expectedObj := range tt.expectedObjects {
				if _, ok := result[expectedObj]; !ok {
					t.Errorf("expected object %s to be in result", expectedObj)
				}
			}
		})
	}
}
