package subscriber

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestCreate(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseSubscribeToMessages := testutils.DataFromFile(t, "create/messages-response-1.json")
	responseSubscribeToMessagesAndUnknown := testutils.DataFromFile(t, "create/partial-success.json")
	responseRollbackDelete := testutils.DataFromFile(t, "create/rollback-success-delete.json")
	errorRollbackDelete := testutils.DataFromFile(t, "create/rollback-fail-delete.json")

	eventTypesCUD := []common.SubscriptionEventType{
		common.SubscriptionEventTypeCreate,
		common.SubscriptionEventTypeUpdate,
		common.SubscriptionEventTypeDelete,
	}

	payloadSubscribeToMessages := []byte(`{
	  "id": "me/messages",
	  "method": "POST",
	  "url": "/subscriptions",
	  "body": {
		"changeType": "created,updated,deleted",
		"notificationUrl": "https://test.com/webhook",
		"resource": "me/messages",
		"clientState": "me/messages"
	  },
	  "headers": {"Content-Type": "application/json"}
	}`)

	payloadSubscribeToMovies := []byte(`{
	  "id": "movies",
	  "method": "POST",
	  "url": "/subscriptions",
	  "body": {
		"changeType": "created",
		"notificationUrl": "https://test.com/webhook",
		"resource": "movies",
		"clientState": "movies"
	  },
	  "headers": {"Content-Type": "application/json"}
	}`)
	payloadRemoveSubscription := []byte(`{
	  "id": "38ca43fa-4602-43ef-a865-aca8a3ddc1ce",
	  "method": "DELETE",
	  "url": "/subscriptions/38ca43fa-4602-43ef-a865-aca8a3ddc1ce",
	  "headers": {"Content-Type": "application/json"}
	}`)

	tests := []testroutines.TestCaseSubscribe{
		{
			Name: "Missing object for subscription",
			Input: common.SubscribeParams{
				Request: &Request{WebhookURL: "https://test.com/webhook"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Invalid subscription request type",
			Input: common.SubscribeParams{
				Request: "invalid",
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"butterflies": {},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{components.ErrInvalidSubscriptionRequestType},
		},
		{
			Name: "Subscription to Microsoft messages and unknown object with rollback",
			Input: common.SubscribeParams{
				Request: &Request{
					WebhookURL: "https://test.com/webhook",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: eventTypesCUD},
					"movies":      {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						// Connector may send request payloads in different order.
						payloadBatchRequests(payloadSubscribeToMessages, payloadSubscribeToMovies),
					},
					Then: mockserver.Response(http.StatusOK, responseSubscribeToMessagesAndUnknown),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(payloadRemoveSubscription),
					},
					Then: mockserver.Response(http.StatusNoContent, responseRollbackDelete),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &Result{},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {},
					"movies":      {},
				},
				Status: "failed",
			},
			ExpectedErrs: []error{testutils.StringError("HTTP status 400: API error in batch response")},
		},
		{
			Name: "Subscription to Microsoft messages and unknown object with failing rollback",
			Input: common.SubscribeParams{
				Request: &Request{
					WebhookURL: "https://test.com/webhook",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: eventTypesCUD},
					"movies":      {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						// Connector may send request payloads in different order.
						payloadBatchRequests(payloadSubscribeToMessages, payloadSubscribeToMovies),
					},
					Then: mockserver.Response(http.StatusOK, responseSubscribeToMessagesAndUnknown),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(payloadRemoveSubscription),
					},
					Then: mockserver.Response(http.StatusOK, errorRollbackDelete),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					Subscriptions: map[string]SubscriptionResource{
						"38ca43fa-4602-43ef-a865-aca8a3ddc1ce": {
							ID:         "38ca43fa-4602-43ef-a865-aca8a3ddc1ce",
							ChangeType: "created,updated,deleted",
							ObjectName: "me/messages",
							WebhookURL: "https://test.com/webhook",
							Resource:   "me/messages",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: eventTypesCUD}, // was created and couldn't clean up
					"movies":      {},                      // was never created
				},
				Status: "failed_to_rollback",
			},
			ExpectedErrs: []error{testutils.StringError("HTTP status 400: API error in batch response")},
		},
		{
			Name: "Subscription to Microsoft messages with success",
			Input: common.SubscribeParams{
				Request: &Request{
					WebhookURL: "https://test.com/webhook",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: eventTypesCUD},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/$batch"),
					payloadBatchRequests(payloadSubscribeToMessages),
				},
				Then: mockserver.Response(http.StatusCreated, responseSubscribeToMessages),
			}.Server(),
			Comparator: testroutines.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					Subscriptions: map[string]SubscriptionResource{
						"5917db81-22d5-426d-af91-e285927592b7": {
							ID:         "5917db81-22d5-426d-af91-e285927592b7",
							ChangeType: "created,updated,deleted",
							ObjectName: "me/messages",
							WebhookURL: "https://test.com/webhook",
							Resource:   "me/messages",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: eventTypesCUD},
				},
				Status: "success",
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableSubscriptionCreator, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}

func resultComparator(expectedResult, actualResult *Result) *testutils.CompareResult {
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
				actualKeys = append(actualKeys, name)
			}
			result.AddDiff("Result.Subscriptions is missing key [%v], but have (%v)",
				key, strings.Join(actualKeys, ","))

			continue
		}

		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].ID", key), expectedValue.ID, actualValue.ID)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].ChangeType", key),
			expectedValue.ChangeType, actualValue.ChangeType)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].ObjectName", key),
			expectedValue.ObjectName, actualValue.ObjectName)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].WebhookURL", key),
			expectedValue.WebhookURL, actualValue.WebhookURL)
		result.Assert(fmt.Sprintf("Result.Subscriptions[%v].Resource", key),
			expectedValue.Resource, actualValue.Resource)
	}

	return result
}

func payloadBatchRequests(requests ...[]byte) mockcond.Condition {
	values := make([]string, len(requests))
	for index, req := range requests {
		values[index] = string(req)
	}

	return mockcond.PermuteJSONBody(`{	"requests": [%requests]}`,
		mockcond.PermuteSlots{{
			Name:     "requests",
			NoQuotes: true,
			Values:   values,
		}},
		mockcond.IgnoreBodyField("expirationDateTime", "requests", "body"),
	)
}

// constructTestStrategy creates a Strategy configured for unit testing.
//
// It uses a mock HTTP client and overrides the base URL to point to a test server.
// A fixed clock is injected to ensure deterministic behavior in tests.
func constructTestStrategy(server *httptest.Server) (*Strategy, error) {
	transport, err := components.NewTransport(providers.Microsoft, common.ConnectorParams{
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestMockServerBaseURL(server.URL)

	client := transport.JSONHTTPClient()
	info := transport.ProviderInfo()
	strategy := NewStrategy(client, info, batch.NewStrategy(client, info))

	return strategy, nil
}
