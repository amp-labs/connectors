package copper

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
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "projects"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Remove Company",
			Input: common.DeleteParams{ObjectName: "companies", RecordId: "73615382"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/developer_api/v1/companies/73615382"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Remove Project",
			Input: common.DeleteParams{ObjectName: "projects", RecordId: "1621193"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/developer_api/v1/projects/1621193"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
