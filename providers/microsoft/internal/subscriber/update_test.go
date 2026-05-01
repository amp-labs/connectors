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

func TestUpdate(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseReadSubscriptions := testutils.DataFromFile(t, "read/subscriptions-response.json")
	responseSubscribeToMessages := testutils.DataFromFile(t, "create/messages.json")
	responseDeleteSubscriptions := testutils.DataFromFile(t, "delete/subscriptions-response.json")
	responseCalendarEventRefresh := testutils.DataFromFile(t, "patch/calendar-events-response.json")
	deleteMessage1 := testutils.DataFromFile(t, "delete/payload-message-1.json")
	deleteMessage2 := testutils.DataFromFile(t, "delete/payload-message-2.json")
	deleteMessage3 := testutils.DataFromFile(t, "delete/payload-message-3.json")
	deleteCalendarEvent := testutils.DataFromFile(t, "delete/payload-calendar-event.json")

	payloadSubscribeToMessages := `{
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
	}`
	payloadRefreshCalendarEvents := `{
	  "id": "me/events",
	  "method": "PATCH",
	  "url": "/subscriptions/63b01115-ba3f-4db6-a1ef-793797ec340a",
	  "body": {
		"expirationDateTime": "2026-03-04T10:00:00Z"
	  },
	  "headers": {
		"Content-Type": "application/json"
	  }
	}`

	tests := []testroutines.UpdateSubscription{
		{
			Name: "Subscription to Outlook messages is replaced, Outlook events are refreshed and extra are removed",
			Input: testroutines.UpdateSubscriptionParams{
				Params: common.SubscribeParams{
					Request: Input{
						WebhookURL: "https://test.com/webhook",
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate, // exists in GET /subscriptions
								common.SubscriptionEventTypeUpdate, // new compared to current GET /subscriptions
								common.SubscriptionEventTypeDelete, // new compared to current GET /subscriptions
							},
						},
						"me/events": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate, // exists in GET /subscriptions
								common.SubscriptionEventTypeDelete, // exists in GET /subscriptions
							},
						},
					},
				},
				PreviousResult: nil,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Fetch current state for subscriptions.
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v1.0/subscriptions"),
					},
					Then: mockserver.Response(http.StatusOK, responseReadSubscriptions),
				}, {
					// Create brand-new subscription to messages.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						mockcond.Body(`{"requests": [` + payloadSubscribeToMessages + `]}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseSubscribeToMessages),
				}, {
					// Subscription to Calendar events should be only refreshed, the expiration time prolonged.
					// This is because the change type for this object doesn't need updating.
					// One subscription is going to expire sooner than the other, we pick the most fresh and refresh it.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						mockcond.Body(`{"requests": [` + payloadRefreshCalendarEvents + `]}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseCalendarEventRefresh),
				}, {
					// Every other subscription to messages does not reflect the desired setup, they will be removed.
					// We have already created the desired subscription in the step above.
					// As for the Calendar events, one subscription is going to expire sooner than the other.
					// The most fresh was already refreshed, the other will be removed.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(deleteMessage1, deleteMessage2, deleteMessage3, deleteCalendarEvent),
					},
					Then: mockserver.Response(http.StatusNoContent, responseDeleteSubscriptions),
				}},
			}.Server(),
			Expected: &common.SubscriptionResult{
				Result: Output{},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
							common.SubscriptionEventTypeDelete,
						},
					},
					"me/events": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeDelete,
						},
					},
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

			tt.Run(t, func() (components.SubscriptionUpdater, error) {
				return constructTestStrategy(tt.Server.URL)
			})
		})
	}
}
