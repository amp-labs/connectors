package stripe

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
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
						"accounts": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
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
						"accounts": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
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
						"accounts": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
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
						"customers": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
								common.SubscriptionEventTypeUpdate,
								common.SubscriptionEventTypeDelete,
							},
						},
						"invoices": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeDelete,
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
							"customers": {
								ID:            "we_123:customers",
								EnabledEvents: []string{"customer.create"},
							},
							"invoices": {
								ID:            "we_123:invoices",
								EnabledEvents: []string{"invoice.created", "invoice.updated"},
							},
						},
					},
				},
			},
			Comparator: testconn.ComparatorSubscriptionWithResult(resultComparator),
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/webhook_endpoints/we_123"),
				},
				Then: mockserver.Response(http.StatusOK, webhookEndpointUpdatedResponse),
			}.Server(),
			Expected: &common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"customers": {
							ID:            "we_123:customers",
							Object:        "webhook_endpoint",
							URL:           "https://webhook.site/test",
							EnabledEvents: []string{"customer.created", "customer.updated", "customer.deleted"},
							Status:        "enabled",
						},
						"invoices": {
							ID:            "we_123:invoices",
							Object:        "webhook_endpoint",
							URL:           "https://webhook.site/test",
							EnabledEvents: []string{"invoice.deleted"},
							Status:        "enabled",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"customers": {
						Events: common.SubscriptionEventTypes{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
							common.SubscriptionEventTypeDelete,
						},
					},
					"invoices": {
						Events: common.SubscriptionEventTypes{
							common.SubscriptionEventTypeDelete,
						},
					},
				},
				Status: "success",
			},
		},
		{
			Name: "Desired state replaces previous objects (removal reconciliation)",
			Input: testconn.UpdateSubscriptionParams{
				Params: common.SubscribeParams{
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"customers": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
								common.SubscriptionEventTypeUpdate,
							},
						},
						"quotes": {
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
						"payment_links": {
							Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
						},
					},
					Result: &SubscriptionResult{
						Subscriptions: map[common.ObjectName]WebhookResponse{
							"payment_links": {
								ID:            "we_123:payment_links",
								EnabledEvents: []string{"payment_link.created"},
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
			Comparator: testconn.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &SubscriptionResult{
					Subscriptions: map[common.ObjectName]WebhookResponse{
						"customers": {
							ID:            "we_123:customers",
							Object:        "webhook_endpoint",
							URL:           "https://webhook.site/test",
							EnabledEvents: []string{"customer.created", "customer.updated"},
							Status:        "enabled",
						},
						"quotes": {
							ID:            "we_123:quotes",
							Object:        "webhook_endpoint",
							URL:           "https://webhook.site/test",
							EnabledEvents: []string{"quote.created"},
							Status:        "enabled",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"customers": {
						Events: common.SubscriptionEventTypes{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
						},
					},
					"quotes": {
						Events: common.SubscriptionEventTypes{
							common.SubscriptionEventTypeCreate,
						},
					},
				},
				Status: "success",
			},
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
func resultComparator(expectedResult, actualResult *SubscriptionResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	if !result.Assert("Result.Subscriptions length",
		len(expectedResult.Subscriptions), len(actualResult.Subscriptions)) {
		return result
	}

	for key, expectedValue := range expectedResult.Subscriptions {
		actualValue, ok := actualResult.Subscriptions[key]
		if !ok {
			actualKeys := make([]string, 0)
			for name := range actualResult.Subscriptions {
				actualKeys = append(actualKeys, name.String())
			}
			result.AddDiff("Result.Subscriptions is missing key [%v], but have (%v)",
				key, strings.Join(actualKeys, ","))

			continue
		}

		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].ID", key), expectedValue.ID, actualValue.ID)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].Object", key),
			expectedValue.Object, actualValue.Object)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].URL", key),
			expectedValue.URL, actualValue.URL)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].EnabledEvents", key),
			expectedValue.EnabledEvents, actualValue.EnabledEvents)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].Status", key),
			expectedValue.Status, actualValue.Status)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].Secret", key),
			expectedValue.Secret, actualValue.Secret)
	}

	return result
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
