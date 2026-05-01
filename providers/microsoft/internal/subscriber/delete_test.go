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

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseReadSubscriptions := testutils.DataFromFile(t, "read/subscriptions-response.json")
	responseDeleteSubscriptions := testutils.DataFromFile(t, "delete/subscriptions-response.json")
	deleteMessage1 := testutils.DataFromFile(t, "delete/payload-message-1.json")
	deleteMessage2 := testutils.DataFromFile(t, "delete/payload-message-2.json")
	deleteMessage3 := testutils.DataFromFile(t, "delete/payload-message-3.json")

	tests := []testroutines.DeleteSubscription{
		{
			Name: "Successfully remove subscriptions to Outlook messages.",
			Input: common.SubscriptionResult{
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v1.0/subscriptions"),
					},
					Then: mockserver.Response(http.StatusOK, responseReadSubscriptions),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v1.0/$batch"),
						payloadBatchRequests(deleteMessage1, deleteMessage2, deleteMessage3),
					},
					Then: mockserver.Response(http.StatusNoContent, responseDeleteSubscriptions),
				}},
			}.Server(),
			Expected:     testroutines.None{},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (components.SubscriptionRemover, error) {
				return constructTestStrategy(tt.Server.URL)
			})
		})
	}
}
