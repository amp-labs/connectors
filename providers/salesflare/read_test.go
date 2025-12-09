package salesflare

import (
	"errors"
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

	responseNotFoundError := testutils.DataFromFile(t, "read/not-found.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "read/contacts/1-first-page.json")
	responseContactsLastPage := testutils.DataFromFile(t, "read/contacts/2-empty-page.json")

	tests := []testroutines.Read{
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("firstname")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFoundError),
			}.Server(),
			ExpectedErrs: []error{
				errors.New("Not Found"),
				common.ErrNotFound,
				common.ErrBadRequest,
			},
		},
		{
			Name: "Read contacts first page incrementally",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "role"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
				Until: time.Date(2025, 1, 1, 0, 0, 0, 0,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contacts"),
					mockcond.QueryParam("modification_after", "2024-09-19T12:30:45.000Z"),
					mockcond.QueryParam("modification_before", "2025-01-01T08:00:00.000Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(276596247),
						"role": "Headhunter",
					},
					Raw: map[string]any{
						"name":   "Jeroen Corthout",
						"email":  "jeroen@salesflare.com",
						"domain": "salesflare.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/contacts?" +
					"limit=10000&" +
					"offset=1&" +
					"modification_after=2024-09-19T12:30:45.000Z&" +
					"modification_before=2025-01-01T08:00:00.000Z",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read contacts last empty page",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/contacts"),
				Then:  mockserver.Response(http.StatusOK, responseContactsLastPage),
			}.Server(),
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

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
