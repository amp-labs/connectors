package subscriber

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	deleteMessage1 := testutils.DataFromFile(t, "delete/message-payload-1.json")
	deleteMessage2 := testutils.DataFromFile(t, "delete/message-payload-2.json")
	deleteMessage3 := testutils.DataFromFile(t, "delete/message-payload-3.json")
	responseDeleteSubscriptions := testutils.DataFromFile(t, "delete/response.json")

	tests := []testroutines.TestCaseDeleteSubscription{
		{
			Name: "Successfully remove subscriptions to Microsoft messages.",
			Input: common.SubscriptionResult{
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {},
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
			Expected:     testroutines.None{},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableSubscriptionRemover, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}
