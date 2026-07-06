package breezy

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []testroutines.TestCaseDelete{
		{
			Name:         "Object name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record id is required",
			Input:        common.DeleteParams{ObjectName: objectPositions},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Unsupported object",
			Input:        common.DeleteParams{ObjectName: "templates", RecordId: "tpl001"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Archive position (no DELETE endpoint)",
			Input: common.DeleteParams{
				ObjectName: objectPositions,
				RecordId:   "pos001",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3/company/" + testCompanyID + "/position/pos001/state"),
					mockcond.Body(`{"state":"archived"}`),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableDeleter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
