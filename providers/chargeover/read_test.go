package chargeover

import (
	_ "embed"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

//go:embed test/customers-page.json
var customersPageResponse []byte

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	chicago := time.FixedZone("America/Chicago", -5*60*60)
	since := time.Date(2026, 10, 5, 10, 18, 49, 0, chicago)
	until := time.Date(2026, 10, 5, 11, 0, 0, 0, chicago)

	// Colons inside a where value must be backslash-escaped, never percent-encoded.
	// See https://developer.chargeover.com/docs/api-information/escape-syntax/example-escape-syntax.
	whereSince := `mod_datetime:GTE:2026-10-05T10\:18\:49-05\:00`
	whereSinceUntil := whereSince + `,mod_datetime:LTE:2026-10-05T11\:00\:00-05\:00`

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Since is sent as a backslash-escaped where clause",
			Input: common.ReadParams{
				ObjectName: "customer",
				Fields:     connectors.Fields("company"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v3/customer"),
					mockcond.QueryParam("where", whereSince),
					mockcond.QueryParam("limit", "500"),
					mockcond.QueryParamsMissing("offset"),
				},
				Then: mockserver.Response(http.StatusOK, customersPageResponse),
				Else: mockserver.ResponseString(http.StatusBadRequest, `{"code":400,"message":"Malformed WHERE clause."}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"company": "Amp LiveTest"},
					Raw:    map[string]any{"customer_id": float64(2), "company": "Amp LiveTest"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Since and Until are combined in one where clause",
			Input: common.ReadParams{
				ObjectName: "customer",
				Fields:     connectors.Fields("company"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v3/customer"),
					mockcond.QueryParam("where", whereSinceUntil),
				},
				Then: mockserver.Response(http.StatusOK, customersPageResponse),
				Else: mockserver.ResponseString(http.StatusBadRequest, `{"code":400,"message":"Malformed WHERE clause."}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"company": "Amp LiveTest"},
					Raw:    map[string]any{"customer_id": float64(2), "company": "Amp LiveTest"},
				}},
				Done: true,
			},
		},
		{
			Name: "Objects without a filtering field ignore Since",
			Input: common.ReadParams{
				ObjectName: "brand",
				Fields:     connectors.Fields("company"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v3/brand"),
					mockcond.QueryParamsMissing("where"),
				},
				Then: mockserver.Response(http.StatusOK, customersPageResponse),
				Else: mockserver.ResponseString(http.StatusBadRequest, `{"code":400,"message":"You cannot filter by: mod_datetime."}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"company": "Amp LiveTest"},
					Raw:    map[string]any{"customer_id": float64(2), "company": "Amp LiveTest"},
				}},
				Done: true,
			},
		},
		{
			Name: "Next page URL is requested as is, where clause intact",
			Input: common.ReadParams{
				ObjectName: "customer",
				Fields:     connectors.Fields("company"),
				Since:      since,
				NextPage:   testconn.URLTestServer + `/api/v3/customer?limit=500&offset=500&where=mod_datetime%3AGTE%3A2026-10-05T10%5C%3A18%5C%3A49-05%5C%3A00`, //nolint:lll
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v3/customer"),
					mockcond.QueryParam("where", whereSince),
					mockcond.QueryParam("offset", "500"),
					mockcond.QueryParam("limit", "500"),
				},
				Then: mockserver.Response(http.StatusOK, customersPageResponse),
				Else: mockserver.ResponseString(http.StatusBadRequest, `{"code":400,"message":"Malformed WHERE clause."}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"company": "Amp LiveTest"},
					Raw:    map[string]any{"customer_id": float64(2), "company": "Amp LiveTest"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestNextRecordsURL(t *testing.T) {
	t.Parallel()

	firstPage, err := urlbuilder.New("https://test.chargeover.com/api/v3/customer")
	if err != nil {
		t.Fatal(err)
	}

	firstPage.WithQueryParam("where", `mod_datetime:GTE:2026-10-05T10\:18\:49-05\:00`)
	firstPage.WithQueryParam("limit", "500")

	t.Run("partial page has no next page", func(t *testing.T) {
		t.Parallel()

		next, err := nextRecordsURL(firstPage, "customer", pageSize-1)(nil)
		if err != nil || next != "" {
			t.Fatalf("expected no next page, got %q, err %v", next, err)
		}
	})

	t.Run("full page advances offset and keeps where clause", func(t *testing.T) {
		t.Parallel()

		next, err := nextRecordsURL(firstPage, "customer", pageSize)(nil)
		if err != nil {
			t.Fatal(err)
		}

		expected := "https://test.chargeover.com/api/v3/customer?limit=500&offset=500" +
			"&where=mod_datetime%3AGTE%3A2026-10-05T10%5C%3A18%5C%3A49-05%5C%3A00"
		if next != expected {
			t.Fatalf("unexpected next page\n got: %s\nwant: %s", next, expected)
		}
	})

	t.Run("objects without pagination never have a next page", func(t *testing.T) {
		t.Parallel()

		next, err := nextRecordsURL(firstPage, "_report", pageSize)(nil)
		if err != nil || next != "" {
			t.Fatalf("expected no next page, got %q, err %v", next, err)
		}
	})
}
