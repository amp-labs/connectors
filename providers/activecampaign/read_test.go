package activecampaign

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrUnprocessable := testutils.DataFromFile(t, "read/err-unprocessable.json")
	responseContactsFirst := testutils.DataFromFile(t, "read/contacts/1-first-page.json")
	responseContactsLast := testutils.DataFromFile(t, "read/contacts/2-last-page.json")

	tests := []testconn.TestCaseRead{
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
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("email")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseErrUnprocessable),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Contact Email Address is not valid."),
			},
		},
		{
			Name:  "Error carrying a bare message is parsed",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("email")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusNotFound,
					`{"message":"No Result found for Subscriber with id 1"}`),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("No Result found for Subscriber with id 1"),
			},
		},
		{
			// ActiveCampaign answers every endpoint with an empty bodied 402 once an account
			// lapses, so the error formats must survive a response with nothing to unmarshal.
			Name:  "Suspended account returns an empty body",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("email")},
			Server: mockserver.Fixed{
				Always: mockserver.ResponseString(http.StatusPaymentRequired, ``),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
		{
			Name: "Read contacts first page",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/contacts"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("orders[id]", "ASC"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"email": "johndoe@example.com"},
					Raw:    map[string]any{"firstName": "John"},
					Id:     "1",
				}, {
					Fields: map[string]any{"email": "janebrown@example.com"},
					Raw:    map[string]any{"firstName": "Jane"},
					Id:     "2",
				}},
				// A page shorter than the requested size means the collection is exhausted.
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Contacts page forward using id_greater",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/contacts"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("orders[id]", "ASC"),
					mockcond.QueryParamsMissing("offset"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsFirst),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: testconn.URLTestServer + "/api/3/contacts?id_greater=2&limit=2&orders%5Bid%5D=ASC",
				Done:     false,
			},
		},
		{
			Name: "Read contacts last page",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				PageSize:   2,
				NextPage: common.NextPageToken(
					testconn.URLTestServer + "/api/3/contacts?id_greater=2&limit=2&orders%5Bid%5D=ASC"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/contacts"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("id_greater", "2"),
					mockcond.QueryParam("orders[id]", "ASC"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsLast),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"email": "adamsmith@example.com"},
					Raw:    map[string]any{"firstName": "Adam"},
					Id:     "3",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Objects lacking id_greater fall back to offset paging",
			Input: common.ReadParams{
				ObjectName: "deals",
				Fields:     connectors.Fields("title"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/deals"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParamsMissing("orders[id]", "id_greater"),
				},
				Then: mockserver.ResponseString(http.StatusOK,
					`{"deals":[{"id":"10","title":"Renewal"}],"meta":{"total":"3"}}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: testconn.URLTestServer + "/api/3/deals?limit=1&offset=1",
				Done:     false,
			},
		},
		{
			Name: "Page size above the provider cap is clamped",
			Input: common.ReadParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("tag"),
				PageSize:   500,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/tags"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"tags":[],"meta":{"total":"0"}}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
		{
			Name: "Incremental read of contacts sends filters[updated_after]",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				Since:      time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/contacts"),
					mockcond.QueryParam("filters[updated_after]", "2024-10-01T00:00:00Z"),
				},
				Then: mockserver.ResponseString(http.StatusOK,
					`{"contacts":[{"id":"3","email":"adamsmith@example.com"}],"meta":{"total":"1"}}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 1, NextPage: "", Done: true},
		},
		{
			Name: "Empty read response is handled",
			Input: common.ReadParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("tag"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/3/tags"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"tags":[],"meta":{"total":"0"}}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: server.Client(),
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(server.URL)

	return connector, nil
}
