package subscriber

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testconn.TestCaseDeleteSubscription{
		{
			Name: "Failer to delete any callback",
			Input: common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						"contacts":        {ID: 0},
						"project/tickets": {ID: 0},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Expected:     testconn.None{},
			ExpectedErrs: []error{common.ErrServer},
		},
		{
			Name: "Successfully remove subscriptions to contacts and tickets.",
			Input: common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						"contacts":        {ID: 26571},
						"project/tickets": {ID: 26572},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks/26571"),
						mockcond.Header(http.Header{"ClientId": []string{"test-client-id"}}),
					},
					Then: mockserver.Response(http.StatusNoContent),
				}, {
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks/26572"),
						mockcond.Header(http.Header{"ClientId": []string{"test-client-id"}}),
					},
					Then: mockserver.Response(http.StatusNoContent),
				}},
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
