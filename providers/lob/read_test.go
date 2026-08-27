package lob

import (
	"net/http"
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

	responseInvalidPageSize := testutils.DataFromFile(t, "read/err-422-page-size-too-large.json")
	responseAddressesFirst := testutils.DataFromFile(t, "read/addresses/1-first-page.json")
	responseAddressesLast := testutils.DataFromFile(t, "read/addresses/2-last-page.json")

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
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "addresses", Fields: connectors.Fields("name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseInvalidPageSize),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("limit must be less than or equal to 100: lob code 422"),
				common.ErrBadRequest,
			},
		},
		{
			Name: "Read addresses first page",
			Input: common.ReadParams{
				ObjectName: "addresses",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/addresses"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseAddressesFirst),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"name": "CHRISTEN GARDNER"},
					Raw:    map[string]any{"email": "Christen@lob.com"},
					Id:     "adr_5c49d278ef3b7986",
				}, {
					Fields: map[string]any{"name": "JOAN LANE"},
					Raw:    map[string]any{"email": "Joan@lob.com"},
					Id:     "adr_317fce961b645606",
				}},
				NextPage: "https://api.lob.com/v1/addresses?limit=2&after=eyJkYXRlT2Zmc2V0IjoiMjAyNi0wOC0xOVQwMToxNDoxMS4yOTJaIiwiaWRPZmZzZXQiOiJhZHJfMzE3ZmNlOTYxYjY0NTYwNiJ9",
				Done:     false,
			},
		},
		{
			Name: "Read addresses last page",
			Input: common.ReadParams{
				ObjectName: "addresses",
				Fields:     connectors.Fields("name"),
				NextPage:   testconn.URLTestServer + "/v1/addresses?limit=2&after=eyJkYXRlT2Zmc2V0IjoiMjAyNi0wOC0xOVQwMToxNDoxMS4yOTJaIiwiaWRPZmZzZXQiOiJhZHJfMzE3ZmNlOTYxYjY0NTYwNiJ9",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/addresses"),
				},
				Then: mockserver.Response(http.StatusOK, responseAddressesLast),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"name": "HARRY ZHANG"},
					Raw:    map[string]any{"email": "harry@lob.com"},
					Id:     "adr_eef4879a15726a16",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read billing_groups with Since",
			Input: common.ReadParams{
				ObjectName: "billing_groups",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/billing_groups"),
					mockcond.QueryParam("date_modified", `{"gt":"2026-05-05T23:10:00Z"}`),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
		{
			Name: "Read billing_groups with Since and Until",
			Input: common.ReadParams{
				ObjectName: "billing_groups",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
				Until:      time.Date(2027, 5, 5, 23, 10, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/billing_groups"),
					mockcond.Or{
						mockcond.QueryParam("date_modified", `{"gt":"2026-05-05T23:10:00Z","lt":"2027-05-05T23:10:00Z"}`),
						mockcond.QueryParam("date_modified", `{"lt":"2027-05-05T23:10:00Z","gt":"2026-05-05T23:10:00Z"}`),
					},
				},
				Then: mockserver.ResponseString(http.StatusOK, `{}`),
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
