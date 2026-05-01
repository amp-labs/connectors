package subscriber

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestCreate(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseSubscribeToMessages := testutils.DataFromFile(t, "create/messages.json")
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
		"clientState": "me/messages",
		"expirationDateTime": "2026-03-04T10:00:00Z"
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
		"clientState": "movies",
		"expirationDateTime": "2026-03-04T10:00:00Z"
	  },
	  "headers": {"Content-Type": "application/json"}
	}`)
	payloadRemoveSubscription := []byte(`{
	  "id": "38ca43fa-4602-43ef-a865-aca8a3ddc1ce",
	  "method": "DELETE",
	  "url": "/subscriptions/38ca43fa-4602-43ef-a865-aca8a3ddc1ce",
	  "headers": {"Content-Type": "application/json"}
	}`)

	tests := []testroutines.CreateSubscription{
		{
			Name: "Missing object for subscription",
			Input: common.SubscribeParams{
				Request: Input{WebhookURL: "https://test.com/webhook"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Invalid subscription request type",
			Input: common.SubscribeParams{
				Request: "invalid",
				SubscriptionEvents: State{
					"butterflies": {},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{components.ErrInvalidSubscriptionRequestType},
		},
		{
			Name: "Subscription to Outlook messages and unknown object with rollback",
			Input: common.SubscribeParams{
				Request: Input{
					WebhookURL: "https://test.com/webhook",
				},
				SubscriptionEvents: State{
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
				ObjectEvents: State{
					"me/messages": {},
					"movies":      {},
				},
				Status: "failed",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Subscription to Outlook messages and unknown object with failing rollback",
			Input: common.SubscribeParams{
				Request: Input{
					WebhookURL: "https://test.com/webhook",
				},
				SubscriptionEvents: State{
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
				ObjectEvents: State{
					"me/messages": {Events: eventTypesCUD}, // was created and couldn't clean up
					"movies":      {},                      // was never created
				},
				Status: "failed_to_rollback",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Subscription to Outlook messages with success",
			Input: common.SubscribeParams{
				Request: Input{
					WebhookURL: "https://test.com/webhook",
				},
				SubscriptionEvents: State{
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
				ObjectEvents: State{
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

			tt.Run(t, func() (components.SubscriptionCreator, error) {
				return constructTestStrategy(tt.Server.URL)
			})
		})
	}
}

func resultComparator(expectedResult, actualResult any) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	// No-op.

	return result
}

func payloadBatchRequests(requests ...[]byte) mockcond.Condition {
	values := make([]string, len(requests))
	for index, req := range requests {
		values[index] = string(req)
	}

	return mockcond.PermuteJSONBody(`{	"requests": [%requests]}`,
		mockcond.PermuteSlot{
			Name:     "requests",
			NoQuotes: true,
			Values:   values,
		},
	)
}
