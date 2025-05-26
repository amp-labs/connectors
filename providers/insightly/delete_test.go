package insightly

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:         "Delete param object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete param object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "Leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "Leads", RecordId: "78572266"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3.1/Leads/78572266"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusAccepted),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
