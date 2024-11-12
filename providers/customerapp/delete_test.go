package customerapp

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

	errorNotFound := testutils.DataFromFile(t, "delete-not-found.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "reporting_webhooks"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:   "Object name is not supported",
			Input:  common.DeleteParams{ObjectName: "customer_exports", RecordId: "95049"},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "reporting_webhooks", RecordId: "95049"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/v1/reporting_webhooks/95049"),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Error on deleting missing record",
			Input: common.DeleteParams{ObjectName: "reporting_webhooks", RecordId: "95049"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"not found (reference 01JCGC85CF663RT1V3FA04ZBNK)",
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
