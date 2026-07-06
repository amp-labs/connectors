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

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	deleteMessage1 := testutils.DataFromFile(t, "delete/message-payload-1.json")
	deleteMessage2 := testutils.DataFromFile(t, "delete/message-payload-2.json")
	deleteMessage3 := testutils.DataFromFile(t, "delete/message-payload-3.json")
	responseDeleteSubscriptions := testutils.DataFromFile(t, "delete/response.json")

	tests := []testconn.TestCaseDeleteSubscription{
		{
			Name: "Successfully remove subscriptions to Microsoft messages.",
			Input: common.SubscriptionResult{
				Result: &Result{
					// Note: The content of the SubscriptionResource is not important for delete.
					Subscriptions: map[string]SubscriptionResource{
						"c27d2493-0518-48db-b994-6d43aa584355": {ObjectName: "me/messages"}, // Message 1
						"29772d64-ee45-4e64-ab82-481602e07bc2": {ObjectName: "me/messages"}, // Message 2
						"90b46999-27c7-4145-a409-a5ef18250522": {ObjectName: "me/messages"}, // Message 3
						"randomid-6caf-4a02-b331-e4685badeea9": {ObjectName: "me/events"},   // Won't be removed.
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {},
					"me/events":   {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/$batch"),
					payloadBatchRequests(deleteMessage1, deleteMessage2, deleteMessage3),
				},
				Then: mockserver.Response(http.StatusNoContent, responseDeleteSubscriptions),
			}.Server(),
			Expected:     testconn.None{},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionRemover, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}
