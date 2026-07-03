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

func TestUpdateSubscription(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	requestWebhookForContacts := testutils.DataFromFile(t, "create/contacts/request.json")
	responseWebhookForContacts := testutils.DataFromFile(t, "create/contacts/response.json")
	requestWebhookForTickets := testutils.DataFromFile(t, "create/tickets/request.json")
	responseWebhookForTickets := testutils.DataFromFile(t, "create/tickets/response.json")
	errorCreateBadRequest1 := testutils.DataFromFile(t, "create/create-webhook-bad-request-1.json")
	deleteWebhookFailed := testutils.DataFromFile(t, "create/remove-webhook-failed.json")

	eventTypesCUD := []common.SubscriptionEventType{
		common.SubscriptionEventTypeCreate,
		common.SubscriptionEventTypeUpdate,
		common.SubscriptionEventTypeDelete,
	}

	tests := []testconn.TestCaseUpdateSubscription{
		{
			Name: "SubscriptionStatusSuccess: Successfully update by creating one and removing one",
			Input: testconn.UpdateSubscriptionParams{
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"project/tickets": {Events: eventTypesCUD},
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result: &Result{
						ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
							"contacts": {
								ID:         26559,
								WebhookURL: "https://test.com/webhook?recordId=",
								ObjectType: "contact",
							},
						},
					},
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"contacts": {Events: eventTypesCUD},
					},
					Status: common.SubscriptionStatusSuccess,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForTickets),
					},
					Then: mockserver.Response(http.StatusOK, responseWebhookForTickets),
				}, {
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks/26559"),
					},
					Then: mockserver.Response(http.StatusNoContent),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareResult),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						"project/tickets": {
							ID:         26552,
							WebhookURL: "https://test.com/webhook?recordId=",
							ObjectType: "ticket",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"project/tickets": {Events: eventTypesCUD},
				},
				Status: common.SubscriptionStatusSuccess,
			},
		},
		{
			Name: "SubscriptionStatusFailed: Create fails (with rollback) and delete succeeds",
			Input: testconn.UpdateSubscriptionParams{
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"contacts":        {Events: eventTypesCUD},
						"project/tickets": {Events: eventTypesCUD},
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result: &Result{
						ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
							"contacts": {
								ID:         26559,
								WebhookURL: "https://test.com/webhook?recordId=",
								ObjectType: "contact",
							},
						},
					},
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{
						"contacts": {Events: eventTypesCUD},
					},
					Status: common.SubscriptionStatusSuccess,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				// contacts is already there, so it won't be in ToCreate.
				// ToCreate will only have project/tickets. For this case it will fail.
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
					mockcond.BodyBytes(requestWebhookForTickets),
				},
				Then: mockserver.Response(http.StatusBadRequest, errorCreateBadRequest1),
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareResult),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						"contacts": {
							ID:         26559,
							WebhookURL: "https://test.com/webhook?recordId=",
							ObjectType: "contact",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts":        {Events: eventTypesCUD},
					"project/tickets": {}, // not created but succeeded to rollback
				},
				Status: common.SubscriptionStatusFailed,
			},
		},
		{
			Name: "SubscriptionStatusFailedToRollback: Create fails and rollback fails",
			Input: testconn.UpdateSubscriptionParams{
				Params: common.SubscribeParams{
					Request: &Request{WebhookURL: "https://test.com/webhook"},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"contacts":        {Events: eventTypesCUD},
						"project/tickets": {Events: eventTypesCUD},
						"activities":      {Events: eventTypesCUD},
					},
				},
				PreviousResult: &common.SubscriptionResult{
					Result:       &Result{},
					ObjectEvents: map[common.ObjectName]common.ObjectEvents{},
					Status:       common.SubscriptionStatusSuccess,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(), // only contacts create will succeed
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForContacts),
					},
					Then: mockserver.Response(http.StatusOK, responseWebhookForContacts),
				}, {
					If: mockcond.And{ // any other create will fail
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
					},
					Then: mockserver.Response(http.StatusBadRequest, errorCreateBadRequest1),
				}, {
					If: mockcond.And{
						mockcond.MethodDELETE(), // rollback will fail
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks/26559"),
					},
					Then: mockserver.Response(http.StatusBadRequest, deleteWebhookFailed),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareResult),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						"contacts": { // couldn't rollback creation.
							ID:         26559,
							WebhookURL: "https://test.com/webhook?recordId=",
							ObjectType: "contact",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					// successful creation and failed to rollback, so it remains
					"contacts":        {Events: eventTypesCUD},
					"project/tickets": {}, // not created
					"activities":      {}, // not created
				},
				Status: common.SubscriptionStatusFailedToRollback,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionUpdater, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}
