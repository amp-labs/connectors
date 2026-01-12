package iterable

import (
	"errors"
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

	responseMissingListErr := testutils.DataFromFile(t, "delete-missing-list-bad-request.json")
	responseList := testutils.DataFromFile(t, "delete-list.json")

	tests := []testroutines.Delete{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "lists"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:   "Cannot remove unknown object",
			Input:  common.DeleteParams{ObjectName: "coupons", RecordId: "132"},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Cannot remove missing list",
			Input: common.DeleteParams{ObjectName: "lists", RecordId: "5052803"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseMissingListErr),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(`error.lists.deleteFailed(5052803)`),
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "lists", RecordId: "5052803"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/api/lists/5052803"),
				},
				Then: mockserver.Response(http.StatusOK, responseList),
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
