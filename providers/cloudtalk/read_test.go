package cloudtalk

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

	responseContacts := testutils.DataFromFile(t, "read/contacts/response.json")
	responseContactsPaginated := testutils.DataFromFile(t, "read/contacts/paginated-response.json")
	responseCallsEmpty := testutils.DataFromFile(t, "read/calls/empty-response.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read contacts successfully",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contacts/index.json"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "123",
						"name":  "John Doe",
						"email": "john.doe@example.com",
					},
					Raw: map[string]any{
						"id":    "123",
						"name":  "John Doe",
						"email": "john.doe@example.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts with pagination",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contacts/index.json"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsPaginated),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "100",
						"name": "Page One",
					},
					Raw: map[string]any{
						"id":   "100",
						"name": "Page One",
					},
				}},
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts with ignored filtering parameters",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name", "email"),
				Since:      time.Now().Add(-1 * time.Hour), // Should be ignored
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contacts/index.json"),
					mockcond.QueryParam("page", "1"),
					// Note: No date_from/date_to params expected here
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "123",
						"name":  "John Doe",
						"email": "john.doe@example.com",
					},
					Raw: map[string]any{
						"id":    "123",
						"name":  "John Doe",
						"email": "john.doe@example.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read calls with date range",
			Input: common.ReadParams{
				ObjectName: "calls",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2023, 11, 14, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/calls/index.json"),
					mockcond.QueryParam("date_from", "2023-11-14 00:00:00"),
					mockcond.QueryParam("date_to", "2023-11-15 00:00:00"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read activity with date range",
			Input: common.ReadParams{
				ObjectName: "activity",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2023, 11, 14, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/activity/index.json"),
					mockcond.QueryParam("date_from", "2023-11-14 00:00:00"),
					mockcond.QueryParam("date_to", "2023-11-15 00:00:00"),
				},
				Then: mockserver.Response(http.StatusOK, responseCallsEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Not found error",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, []byte(`{"error": "Not Found"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
			},
		},
		{
			Name: "Read Internal server error",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, []byte(`{"error": "Internal Server Error"}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrServer,
				errors.New("Internal Server Error"), //nolint:goerr113
			},
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
