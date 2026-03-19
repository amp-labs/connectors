package fourfour

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")
	leadsResponse := testutils.DataFromFile(t, "leads.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of contacts",
			Input: common.ReadParams{ObjectName: "Contacts", Fields: connectors.Fields("id", "first_name", "email")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/odata/Contacts"),
				Then:  mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "string",
						"first_name": "string",
						"email":      "string",
					},
					Raw: map[string]any{
						"last_name":   "string",
						"phone":       "string",
						"title":       "string",
						"account_id":  "string",
						"department":  "string",
						"lead_source": "string",
						"owner_id":    "string",
						"region":      "string",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of leads sends $filter with updated field",
			Input: common.ReadParams{
				ObjectName: "Leads",
				Fields:     connectors.Fields("id", "email", "company"),
				Since:      time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/odata/Leads"),
					mockcond.QueryParam("$filter", "updated ge 2024-01-15T10:30:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, leadsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":      "string",
						"email":   "string",
						"company": "string",
					},
					Raw: map[string]any{
						"first_name":  "string",
						"last_name":   "string",
						"lead_source": "string",
						"status":      "string",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Objects not in objectSinceField do not send $filter",
			Input: common.ReadParams{
				ObjectName: "Conversations",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/odata/Conversations"),
					mockcond.QueryParamsMissing("$filter"),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
