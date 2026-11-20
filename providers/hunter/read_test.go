package hunter

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	leadsResponse := testutils.DataFromFile(t, "leads-read.json")
	accountResponse := testutils.DataFromFile(t, "account-read.json")

	tests := []testconn.TestCaseRead{
		{
			Name:  "Successfully read leads",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("id", "email")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/leads"),
				Then:  mockserver.Response(http.StatusOK, leadsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(196271291),
						"email": "jane.doe@example.com",
					},
					Raw: map[string]any{
						"first_name": "Jane",
						"company":    "Example",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully read singleton account",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("email", "plan_name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/account"),
				Then:  mockserver.Response(http.StatusOK, accountResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email":     "jane.doe@example.com",
						"plan_name": "Free",
					},
					Raw: map[string]any{
						"first_name": "Jane",
						"team_id":    float64(4841347),
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Singleton account with empty data yields no records",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("email")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/account"),
				Then:  mockserver.Response(http.StatusOK, []byte(`{"data": {}}`)),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
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

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
