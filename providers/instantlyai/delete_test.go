package instantlyai

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

	errorNotFound := testutils.DataFromFile(t, "delete-missing-custom-tags.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "custom-tags"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:   "Object name is not supported",
			Input:  common.DeleteParams{ObjectName: "background-jobs", RecordId: "675266e304a8e55b17f0228b"},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "custom-tags", RecordId: "8cba2d22-cdd0-453f-ac69-4c43f71942e5"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/v2/custom-tags/8cba2d22-cdd0-453f-ac69-4c43f71942e5"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Error on deleting missing record",
			Input: common.DeleteParams{ObjectName: "custom-tags", RecordId: "4fa0ca9d-f205-4a68-8756-5ad62123a53a"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					" Not Found",
				),
			},
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
