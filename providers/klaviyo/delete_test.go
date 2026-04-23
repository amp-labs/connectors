package klaviyo

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

	errorNotFound := testutils.DataFromFile(t, "delete-tag-not-found.json")

	header := http.Header{"revision": []string{"2024-10-15"}}

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "tags"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:   "Object name is not supported",
			Input:  common.DeleteParams{ObjectName: "orders", RecordId: "7da3b722-7e43-55ec-8450-4247843970ab"},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "tags", RecordId: "7da3b722-7e43-55ec-8450-4247843970ab"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/vnd.api+json"),
				If: mockcond.And{
					mockcond.Path("/api/tags/7da3b722-7e43-55ec-8450-4247843970ab"),
					mockcond.MethodDELETE(),
					mockcond.Header(header),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Error on deleting missing record",
			Input: common.DeleteParams{ObjectName: "tags", RecordId: "7da3b722-7e43-55ec-8450-4247843970ab"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME("application/vnd.api+json"),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError(
					"Not found: A tag with id 5eb337d5-a132-4627-aa1e-04bc9aac260d does not exist.",
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
