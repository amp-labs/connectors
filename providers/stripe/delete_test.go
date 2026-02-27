package stripe

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorNotFound := testutils.DataFromFile(t, "delete/err-not-found.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "customers", RecordId: "cus_Rd2zeYCh5b9gtC"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v1/customers/cus_Rd2zeYCh5b9gtC"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Error on deleting missing record",
			Input: common.DeleteParams{ObjectName: "customers", RecordId: "cus_Rd2zeYCh5b9gtC"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v1/customers/cus_Rd2zeYCh5b9gtC"),
				},
				Then: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError(
					"No such customer: 'cus_Rd2zeYCh5b9gtC'",
				),
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
