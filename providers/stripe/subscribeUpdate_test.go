package stripe

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
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
						WebhookId: "we_123",
						Secret:    "secret_7",
						Subscriptions: map[common.ObjectName][]string{
							"customers": {"customer.create"},
							"invoices":  {"invoice.created", "invoice.updated"},
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
					WebhookId: "we_123",
					Secret:    "secret_7",
					Subscriptions: map[common.ObjectName][]string{
						"customers": {"customer.created", "customer.updated", "customer.deleted"},
						"invoices":  {"invoice.deleted"},
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
						WebhookId: "we_123",
						Secret:    "secret_7",
						Subscriptions: map[common.ObjectName][]string{
							"payment_links": {"payment_link.created"},
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
					WebhookId: "we_123",
					Secret:    "secret_7",
					Subscriptions: map[common.ObjectName][]string{
						"customers": {"customer.created", "customer.updated"},
						"quotes":    {"quote.created"},
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
	result.Assert("Result.WebhookId", expectedResult.WebhookId, actualResult.WebhookId)
	result.Assert("Result.Secret", expectedResult.Secret, actualResult.Secret)

	if !result.Assert("Result.Subscriptions length",
		len(expectedResult.Subscriptions), len(actualResult.Subscriptions)) {
		return result
	}

	for key, expectedEvents := range expectedResult.Subscriptions {
		actualEvents, ok := actualResult.Subscriptions[key]
		if !ok {
			actualKeys := make([]string, 0)
			for name := range actualResult.Subscriptions {
				actualKeys = append(actualKeys, name.String())
			}
			result.AddDiff("Result.Subscriptions is missing key [%v], but have (%v)",
				key, strings.Join(actualKeys, ","))

			continue
		}

		sort.Strings(expectedEvents)
		sort.Strings(actualEvents)

		result.Assert(fmt.Sprintf("Result.Subscriptions[%v]", key), expectedEvents, actualEvents)
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
					WebhookId: "we_123",
					Subscriptions: map[common.ObjectName][]string{
						"accounts": {"account.updated"},
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
