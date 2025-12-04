package dropboxsign

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	bulkJobsResponse := testutils.DataFromFile(t, "read-bulk_send_job.json")
	faxResponse := testutils.DataFromFile(t, "read-fax.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successful read of customers with chosen fields",
			Input: common.ReadParams{ObjectName: "customers", Fields: connectors.Fields("id", "name", "email", "status")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/customers"),
				Then:  mockserver.Response(http.StatusOK, bulkJobsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
						"status": "active",
						"name":   "Jo Brown-Anderson",
						"email":  "jo@example.com",
					},
					Raw: map[string]any{
						"id":                "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
						"status":            "active",
						"custom_data":       nil,
						"name":              "Jo Brown-Anderson",
						"email":             "jo@example.com",
						"marketing_consent": false,
						"locale":            "en",
						"created_at":        "2024-04-11T15:57:24.813Z",
						"updated_at":        "2024-04-11T15:59:56.658719Z",
						"import_meta":       nil,
					},
				}},
				NextPage: "https://api.paddle.com/customers?after=ctm_01h8441jn5pcwrfhwh78jqt8hk",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page for customers",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "name", "email", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/customers"),
				Then:  mockserver.Response(http.StatusOK, faxResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
