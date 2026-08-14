package stripe

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestUpdateSubscription(t *testing.T) { // nolint:funlen
	t.Parallel()

	webhookEndpointUpdatedResponse := testutils.DataFromFile(t, "subscribe/webhook-endpoint-updated-response.json")

	tests := []testconn.TestCaseUpdateSubscription{
		{
			Name: "Missing previous result",
			Input: testconn.UpdateSubscriptionParams{
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
			Input: testconn.UpdateSubscriptionParams{
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
			Input: testconn.UpdateSubscriptionParams{
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
			Input: testconn.UpdateSubscriptionParams{
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
			Comparator:   compareUpdateResultObjects("account"),
		},
		{
			Name: "Desired state replaces previous objects (removal reconciliation)",
			Input: testconn.UpdateSubscriptionParams{
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
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"balance": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
						},
					},
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
			// "balance" is absent from the desired state, so it must not survive the update.
			Comparator: compareUpdateResultObjects("account", "charge"),
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionUpdater, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

// compareUpdateResultObjects verifies the update succeeded and that the resulting state
// contains exactly the desired objects (both in ObjectEvents and stored Subscriptions).
func compareUpdateResultObjects(
	objects ...common.ObjectName,
) func(string, *common.SubscriptionResult, *common.SubscriptionResult) *testutils.CompareResult {
	return func(_ string, actual, _ *common.SubscriptionResult) *testutils.CompareResult {
		result := testutils.NewCompareResult()

		if actual == nil {
			return result.AddDiff("subscription result is nil")
		}

		result.Assert("Status", common.SubscriptionStatusSuccess, actual.Status)
		result.Assert("ObjectEvents length", len(objects), len(actual.ObjectEvents))

		subscriptionData, ok := actual.Result.(*SubscriptionResult)
		if !ok {
			return result.AddDiff("expected Result of type *SubscriptionResult, got %T", actual.Result)
		}

		result.Assert("Subscriptions length", len(objects), len(subscriptionData.Subscriptions))

		for _, obj := range objects {
			if _, found := actual.ObjectEvents[obj]; !found {
				result.AddDiff("ObjectEvents is missing object [%v]", obj)
			}

			if _, found := subscriptionData.Subscriptions[obj]; !found {
				result.AddDiff("Subscriptions is missing object [%v]", obj)
			}
		}

		return result
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
