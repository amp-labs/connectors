package nutshell

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

	errorServerErrorJSON := testutils.DataFromFile(t, "err-internal-server.json")
	errorServerErrorTXT := testutils.DataFromFile(t, "err-internal-server.txt")
	errorNotFoundHTML := testutils.DataFromFile(t, "err-not-found.html")
	errorPageSize := testutils.DataFromFile(t, "read/err-page-size.json")
	responseAccountsFirstPage := testutils.DataFromFile(t, "read/accounts/1-first-page.json")
	responseAccountsEmptyPage := testutils.DataFromFile(t, "read/accounts/2-empty-page.json")
	responseEvents := testutils.DataFromFile(t, "read/events/1-first-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error with short uninformative JSON format",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, errorServerErrorJSON),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrServer,
				errors.New("Internal Server Error"),
			},
		},
		{
			Name:  "Error with short uninformative text format",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentText(),
				// The status is indeed 405, while the text is internal error.
				Always: mockserver.Response(http.StatusMethodNotAllowed, errorServerErrorTXT),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(`"Internal Error"`),
			},
		},
		{
			Name:  "Error with short uninformative text format",
			Input: common.ReadParams{ObjectName: "products", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFoundHTML),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				errors.New("An Error Occurred: Not Found"),
			},
		},
		{
			Name:  "Error with standard format",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorPageSize),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Invalid filter [Invalid page number: -1]"),
			},
		},
		{
			Name: "Read accounts first page",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("name"),
				Since:      time.Unix(1754518014, 0),
				Until:      time.Unix(1754518016, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/accounts"),
					mockcond.QueryParam("page[limit]", "10000"),
					mockcond.QueryParam("filter[createdTime]", "2025-08-06T22:06:54Z 2025-08-06T22:06:56Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Apple",
					},
					Raw: map[string]any{
						"id":   "15-accounts",
						"type": "accounts",
						"href": "https://app.nutshell.com/rest/accounts/15-accounts",
					},
				}, {
					Fields: map[string]any{
						"name": "Banana",
					},
					Raw: map[string]any{
						"id":   "11-accounts",
						"type": "accounts",
						"href": "https://app.nutshell.com/rest/accounts/11-accounts",
					},
				}, {
					Fields: map[string]any{
						"name": "Cucumber",
					},
					Raw: map[string]any{
						"id":   "19-accounts",
						"type": "accounts",
						"href": "https://app.nutshell.com/rest/accounts/19-accounts",
					},
				}},
				NextPage: testroutines.URLTestServer + "/rest/accounts?" +
					"filter[createdTime]=2025-08-06T22:06:54Z 2025-08-06T22:06:56Z&" +
					"page[limit]=10000&" +
					"page[page]=1",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read accounts second page as last page",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("name"),
				NextPage: testroutines.URLTestServer + "/rest/accounts?" +
					"filter[createdTime]=2025-08-06T22:06:54Z 2025-08-06T22:06:56Z&page[limit]=10000&page[page]=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/accounts"),
					mockcond.QueryParam("page[limit]", "10000"),
					mockcond.QueryParam("page[page]", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountsEmptyPage),
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
			Name: "Read events which has unique response format and pagination",
			Input: common.ReadParams{
				ObjectName: "events",
				Fields:     connectors.Fields("id"),
				Since:      time.Unix(1754518014, 0),
				Until:      time.Unix(1754518016, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/events"),
					mockcond.QueryParam("limit", "10000"),
					mockcond.QueryParam("since_time", "1754518014"),
					mockcond.QueryParam("max_time", "1754518016"),
					mockcond.QueryParamsMissing("page[limit]"), // "events" endpoint uses different query param.
				},
				Then: mockserver.Response(http.StatusOK, responseEvents),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "163-events",
					},
					Raw: map[string]any{
						"actorType":   "users",
						"payloadType": "accounts",
						"action":      "create",
					},
				}, {
					Fields: map[string]any{
						"id": "39-events",
					},
					Raw: map[string]any{
						"actorType":   "origins",
						"payloadType": "leads",
						"action":      "stagechange",
					},
				}, {
					Fields: map[string]any{
						"id": "19-events",
					},
					Raw: map[string]any{
						"actorType":   "origins",
						"payloadType": "contacts",
						"action":      "create",
					},
				}},
				NextPage: "https://app.nutshell.com/rest/events?limit=20&since_id=163-events&since_time=1755182580",
				Done:     false,
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
