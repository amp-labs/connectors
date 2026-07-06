package subscriber

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestUpdate(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	type events = []common.SubscriptionEventType
	createType := common.SubscriptionEventTypeCreate
	updateType := common.SubscriptionEventTypeUpdate
	deleteType := common.SubscriptionEventTypeDelete

	// Inputs for the test: payloads and expected HTTP responses used by the mock server.
	// - create: payloads used when requesting new subscriptions (messages + groups).
	payloadSubscribeToMessages1 := testutils.DataFromFile(t, "create/messages-payload-1.json")
	payloadSubscribeToMessages2 := testutils.DataFromFile(t, "create/messages-payload-2.json")
	payloadSubscribeToGroups := testutils.DataFromFile(t, "create/groups-payload.json")
	responseSubscribeToMessagesAndGroups := testutils.DataFromFile(t, "create/messages-and-groups-response.json")
	responseSubscribeToMessages1 := testutils.DataFromFile(t, "create/messages-response-1.json")
	responseSubscribeToMessages2 := testutils.DataFromFile(t, "create/messages-response-2.json")
	// - patch/refresh: payload used to refresh/extend an existing calendar/events subscription.
	payloadCalendarEventsRefresh := testutils.DataFromFile(t, "patch/calendar-events-payload.json")
	responseCalendarEventRefresh := testutils.DataFromFile(t, "patch/calendar-events-response.json")
	// - delete: payloads used to remove obsolete subscriptions (chat, old message subscription).
	deleteChat := testutils.DataFromFile(t, "delete/chat-payload.json")
	deleteMessage1 := testutils.DataFromFile(t, "delete/message-payload-1.json")
	deleteMessage2 := testutils.DataFromFile(t, "delete/message-payload-2.json")
	responseDelete := testutils.DataFromFile(t, "delete/response.json")

	// List of identifiers used as Inputs or Outputs for the unit tests below.
	// The ids used in output will be found in the JSON files under "test" repo.
	idMessage1 := "c27d2493-0518-48db-b994-6d43aa584355"            // Input: Yes. Output: N/A.
	idMessage2 := "29772d64-ee45-4e64-ab82-481602e07bc2"            // Input: Yes. Output: N/A.
	idMessageReplaceToCRD := "5917db81-22d5-426d-af91-e285927592b7" // Input: N/A. Output: (create,update,delete)
	idMessageReplaceToCD := "813ef41d-28dd-4bc8-a5d3-10bdfc033c99"  // Input: N/A. Output: (create,delete)
	idCalendarEvent := "63b01115-ba3f-4db6-a1ef-793797ec340a"       // Input: Yes. Output: (created,deleted)
	idChat := "a33272c7-b187-444c-840f-9e78cc6de127"                // Input: Yes. Output: N/A.
	idGroupBatchResponse := "9d94e78b-4bca-4fd6-b5f4-bc5f26de5bcf"  // Input: N/A. Output: (deleted)

	tests := []testconn.TestCaseUpdateSubscription{
		{
			// =========================================================================
			// TEST 1: Mix of creating, updating, refreshing and removing Microsoft Graph events
			// =========================================================================
			// This test validates the complete subscription reconciler workflow:
			// - Creating new subscriptions for messages (with upgraded event types) and groups
			// - Refreshing an existing calendar/events subscription (extending expiration)
			// - Deleting obsolete subscriptions (old messages subscription + leftover chat subscription)
			Name: "Mix of creating, updating, refreshing and removing Microsoft Graph events",
			Input: testconn.UpdateSubscriptionParams{
				PreviousResult: &common.SubscriptionResult{
					// PreviousResult simulates the current state known to the connector.
					// It contains several existing subscriptions and respective event types already tracked.
					Result: &Result{
						Subscriptions: map[string]SubscriptionResource{
							// Existing subscription to messages - expected to be replaced by a new
							// subscription that supports created, updated and deleted events.
							idMessage1: {
								ID:         idMessage1,
								ObjectName: "me/messages",
								ChangeType: "created",
								Resource:   "me/messages",
								WebhookURL: "https://test.com/webhook",
							},
							// Existing calendar/events subscription tracking created and deleted.
							// It should be refreshed (expiration extended).
							idCalendarEvent: {
								ID:         idCalendarEvent,
								ObjectName: "me/events",
								ChangeType: "created,deleted",
								Resource:   "me/events",
								WebhookURL: "https://test.com/webhook",
							},
							// Leftover chat subscription present previously but not requested anymore.
							// Expected to be removed.
							idChat: {
								ID:         idChat,
								ObjectName: "chats",
								ChangeType: "created",
								Resource:   "chats",
								WebhookURL: "https://test.com/webhook",
							},
						},
					},
					// ObjectEvents defines which event types were being tracked per object
					// in the previous result. The connector uses these to decide create/update/delete logic.
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {Events: events{createType}},
						"me/events":   {Events: events{createType, deleteType}},
						"chats":       {Events: events{createType}},
					},
					Status: common.SubscriptionStatusSuccess,
				},
				// Params represents desired subscription state after update.
				// It contains objects that should be created, refreshed, updated or deleted.
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": { // Upgrade messages subscription: add update and delete events.
							Events: events{createType, updateType, deleteType},
						},
						"me/events": {Events: events{createType, deleteType}}, // No change in types: refresh only.
						"groups":    {Events: events{deleteType}},             // New object: create subscription.
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Case 1: Create subscriptions for messages and groups in a single batch request.
					// The test expects POST /v1.0/$batch with the combined payload for new subscriptions.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(payloadSubscribeToMessages1, payloadSubscribeToGroups),
					},
					Then: mockserver.Response(http.StatusCreated, responseSubscribeToMessagesAndGroups),
				}, {
					// Case 2: Refresh existing calendar/events subscription expiration (PATCH).
					// No change to changeType is supported by API; we only prolong the subscription.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(payloadCalendarEventsRefresh),
					},
					Then: mockserver.Response(http.StatusCreated, responseCalendarEventRefresh),
				}, {
					// Case 3: Delete obsolete subscriptions:
					// - An old messages subscription is removed because a newer one was created ("update").
					// - The chat subscription is removed because it's no longer requested (left-over).
					// Both deletes are expected in a single batch request.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(deleteMessage1, deleteChat),
					},
					Then: mockserver.Response(http.StatusNoContent, responseDelete),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				// Expected state after reconciler runs:
				// - New messages subscription replaces the old one and supports created,updated,deleted.
				// - events subscription remains but gets its expiration refreshed.
				// - groups subscription is newly created for delete events.
				Result: &Result{
					Subscriptions: map[string]SubscriptionResource{
						idMessageReplaceToCRD: {
							ID:         idMessageReplaceToCRD,
							ObjectName: "me/messages",
							Resource:   "me/messages",
							ChangeType: "created,updated,deleted",
							WebhookURL: "https://test.com/webhook",
						},
						idCalendarEvent: {
							ID:         idCalendarEvent,
							ObjectName: "me/events",
							ChangeType: "created,deleted",
							Resource:   "me/events",
							WebhookURL: "https://test.com/webhook",
						},
						idGroupBatchResponse: {
							ID:         idGroupBatchResponse,
							ObjectName: "groups",
							ChangeType: "deleted",
							Resource:   "groups",
							WebhookURL: "https://test.com/webhook",
						},
					},
				},
				// ObjectEvents after update: chats becomes empty (removed), others reflect requested types.
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: events{createType, updateType, deleteType}},
					"me/events":   {Events: events{createType, deleteType}},
					"groups":      {Events: events{deleteType}},
					"chats":       {Events: nil},
				},
				Status: "success",
			},
			ExpectedErrs: nil,
		},
		{
			// =========================================================================
			// TEST 2: SubscriptionStatusFailed - Refresh fails
			// =========================================================================
			// This test validates error handling when refreshing a subscription fails:
			// - Only one existing subscription (me/events) that needs to be refreshed
			// - Mock server returns 500 Internal Server Error on the refresh request
			// - Expected result: SubscriptionStatusFailed with common.ErrServer
			// - Original subscription remains intact.
			Name: "SubscriptionStatusFailed: Refresh fails",
			Input: testconn.UpdateSubscriptionParams{
				PreviousResult: &common.SubscriptionResult{
					Result: &Result{
						Subscriptions: map[string]SubscriptionResource{
							idCalendarEvent: {
								ID:         idCalendarEvent,
								ObjectName: "me/events",
								ChangeType: "created,deleted",
								Resource:   "me/events",
								WebhookURL: "https://test.com/webhook",
							},
						},
					},
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"me/events": {Events: events{createType, deleteType}},
					},
					Status: common.SubscriptionStatusSuccess,
				},
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/events": {Events: events{createType, deleteType}},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				// Mock server expects exactly one PATCH request to refresh the calendar/events subscription
				// and returns 500 Internal Server Error to simulate refresh failure.
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/$batch"),
					payloadBatchRequests(payloadCalendarEventsRefresh),
				},
				Then: mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					Subscriptions: map[string]SubscriptionResource{
						idCalendarEvent: {
							ID:         idCalendarEvent,
							ObjectName: "me/events",
							ChangeType: "created,deleted",
							Resource:   "me/events",
							WebhookURL: "https://test.com/webhook",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/events": {Events: events{createType, deleteType}},
				},
				Status: common.SubscriptionStatusFailed,
			},
			ExpectedErrs: []error{common.ErrServer},
		},
		{
			// =========================================================================
			// TEST 3: SubscriptionStatusFailed - Delete fails during update
			// =========================================================================
			// This test validates error handling when deleting an obsolete subscription fails:
			// - Existing subscription to messages (created only) needs to be upgraded (created,updated,deleted)
			// - Case 1: Create new subscription succeeds (201 Created)
			// - Case 2: Delete old subscription fails (500 Internal Server Error)
			// - Expected result: SubscriptionStatusFailed with common.ErrServer
			// - Both subscriptions remain: new one (updated events) + old one (delete failed)
			Name: "SubscriptionStatusFailed: Delete fails during update",
			Input: testconn.UpdateSubscriptionParams{
				PreviousResult: &common.SubscriptionResult{
					Result: &Result{
						Subscriptions: map[string]SubscriptionResource{
							idMessage1: {
								ID:         idMessage1,
								ObjectName: "me/messages",
								ChangeType: "created",
								Resource:   "me/messages",
								WebhookURL: "https://test.com/webhook",
							},
						},
					},
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {Events: events{createType}},
					},
					Status: common.SubscriptionStatusSuccess,
				},
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {Events: events{createType, updateType, deleteType}},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Case 1: Create new messages subscription with upgraded event types succeeds.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(payloadSubscribeToMessages1),
					},
					Then: mockserver.Response(http.StatusCreated, responseSubscribeToMessages1),
				}, {
					// Case 2: Delete old messages subscription fails with 500 error.
					// This simulates a transient network issue or API error during deletion.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(deleteMessage1, deleteMessage2),
					},
					Then: mockserver.Response(http.StatusInternalServerError),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					Subscriptions: map[string]SubscriptionResource{
						idMessage1: {
							ID:         idMessage1,
							ObjectName: "me/messages",
							ChangeType: "created",
							Resource:   "me/messages",
							WebhookURL: "https://test.com/webhook",
						},
						idMessageReplaceToCRD: {
							ID:         idMessageReplaceToCRD,
							ObjectName: "me/messages",
							Resource:   "me/messages",
							ChangeType: "created,updated,deleted",
							WebhookURL: "https://test.com/webhook",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: events{createType, updateType, deleteType}},
				},
				Status: common.SubscriptionStatusFailed,
			},
			ExpectedErrs: []error{common.ErrServer},
		},
		{
			// =========================================================================
			// TEST 4: Previous result has 2 subscriptions for one object which is cleaned up
			// =========================================================================
			// This test validates cleanup logic when duplicate subscriptions exist for the same object:
			// - Previous state has TWO subscriptions for "me/messages":
			//   1. subscription A: created,updated,deleted (will be replaced with created,deleted)
			//   2. subscription B: created only (will be removed)
			// - Desired state: only created,deleted events for messages
			// - Case 1: Create new subscription with created,deleted succeeds (201 Created)
			// - Case 2: Delete BOTH old subscriptions (A + B) succeeds (204 No Content)
			// - Expected result: SubscriptionStatusSuccess with only the new subscription remaining
			Name: "Previous result has 2 subscriptions for one object which is cleaned up",
			Input: testconn.UpdateSubscriptionParams{
				PreviousResult: &common.SubscriptionResult{
					Result: &Result{
						Subscriptions: map[string]SubscriptionResource{
							idMessage1: {
								ID:         idMessage1,
								ObjectName: "me/messages",
								ChangeType: "created", // will be removed
								Resource:   "me/messages",
								WebhookURL: "https://test.com/webhook",
							},
							idMessage2: {
								ID:         idMessage2,
								ObjectName: "me/messages",
								Resource:   "me/messages",
								ChangeType: "created,updated,deleted", // will be replaced with created,deleted
								WebhookURL: "https://test.com/webhook",
							},
						},
					},
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {Events: events{createType, updateType, deleteType}},
					},
					Status: common.SubscriptionStatusSuccess,
				},
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {Events: events{createType, deleteType}},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Case 1: Create new messages subscription with reduced event types (created,deleted).
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(payloadSubscribeToMessages2),
					},
					Then: mockserver.Response(http.StatusCreated, responseSubscribeToMessages2),
				}, {
					// Case 2: Delete BOTH obsolete subscriptions in a single batch request:
					// - subscription A (created,updated,deleted) - replaced by new one
					// - subscription B (created only) - no longer needed
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(deleteMessage1, deleteMessage2),
					},
					Then: mockserver.Response(http.StatusNoContent, responseDelete),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(resultComparator),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					Subscriptions: map[string]SubscriptionResource{
						idMessageReplaceToCD: {
							ID:         idMessageReplaceToCD,
							ObjectName: "me/messages",
							Resource:   "me/messages",
							ChangeType: "created,deleted",
							WebhookURL: "https://test.com/webhook",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {Events: events{createType, deleteType}},
				},
				Status: common.SubscriptionStatusSuccess,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionUpdater, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}
