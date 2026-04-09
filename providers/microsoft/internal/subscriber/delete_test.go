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

	_ = testutils.DataFromFile(t, "delete-not-found.json")

	tests := []testroutines.DeleteSubscription{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Write object and its ID must be included",
			Input: common.SubscriptionResult{
				Result:            nil,
				ObjectEvents:      nil,
				Status:            "",
				Objects:           nil,
				Events:            nil,
				UpdateFields:      nil,
				PassThroughEvents: nil,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "",
			Input: common.SubscriptionResult{
				Result:            nil,
				ObjectEvents:      nil,
				Status:            "",
				Objects:           nil,
				Events:            nil,
				UpdateFields:      nil,
				PassThroughEvents: nil,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
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
