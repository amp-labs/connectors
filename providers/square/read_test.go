package square

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

func TestRead(t *testing.T) {
	t.Parallel()

	customersResponse := testutils.DataFromFile(t, "customers.json")
	customersLastPage := testutils.DataFromFile(t, "customers-last-page.json")
	locationsResponse := testutils.DataFromFile(t, "locations.json")
	paymentsResponse := testutils.DataFromFile(t, "payments.json")
	payoutsResponse := testutils.DataFromFile(t, "payouts.json")
	merchantsResponse := testutils.DataFromFile(t, "merchants.json")
	giftCardsResponse := testutils.DataFromFile(t, "gift_cards.json")
	cardsResponse := testutils.DataFromFile(t, "cards.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read needs an object",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Read needs at least one field",
			Input:        common.ReadParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Unknown object is not supported",
			Input: common.ReadParams{
				ObjectName: "butterflies",
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name: "Read customers first page with limit and next-page cursor",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "given_name"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/customers"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, customersResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "JDKYHBWT1D4F8MFH63DBMEN8Y4",
						Fields: map[string]any{
							"id":         "JDKYHBWT1D4F8MFH63DBMEN8Y4",
							"given_name": "Amelia",
						},
						Raw: map[string]any{
							"id":            "JDKYHBWT1D4F8MFH63DBMEN8Y4",
							"given_name":    "Amelia",
							"family_name":   "Earhart",
							"email_address": "Amelia.Earhart@example.com",
						},
					},
				},
				NextPage: "GcZjJVTwYth6PnqWQQHwx",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read customers second page defaults limit to the max and forwards the cursor",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "given_name"),
				NextPage:   "GcZjJVTwYth6PnqWQQHwx",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/customers"),
					mockcond.QueryParam("cursor", "GcZjJVTwYth6PnqWQQHwx"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, customersLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "V9PD6FN1KMSQ8MEEYQHK0X1AC4",
						Fields: map[string]any{
							"id":         "V9PD6FN1KMSQ8MEEYQHK0X1AC4",
							"given_name": "Wilbur",
						},
						Raw: map[string]any{
							"id":          "V9PD6FN1KMSQ8MEEYQHK0X1AC4",
							"given_name":  "Wilbur",
							"family_name": "Wright",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read locations is a single non-paginated page",
			Input: common.ReadParams{
				ObjectName: "locations",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/locations"),
				Then:  mockserver.Response(http.StatusOK, locationsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "18YC4JDH91E1H",
						Fields: map[string]any{
							"id":   "18YC4JDH91E1H",
							"name": "Default Test Account",
						},
						Raw: map[string]any{
							"id":     "18YC4JDH91E1H",
							"name":   "Default Test Account",
							"status": "ACTIVE",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read payments with Since and Until adds begin_time and end_time",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("id", "status"),
				Since:      time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2021, 10, 31, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/payments"),
					mockcond.QueryParam("begin_time", "2021-10-01T00:00:00Z"),
					mockcond.QueryParam("end_time", "2021-10-31T23:59:59Z"),
				},
				Then: mockserver.Response(http.StatusOK, paymentsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "bP9mAxakQF1Fp7gM92R1yk1ScjMZY",
						Fields: map[string]any{
							"id":     "bP9mAxakQF1Fp7gM92R1yk1ScjMZY",
							"status": "COMPLETED",
						},
						Raw: map[string]any{
							"id":          "bP9mAxakQF1Fp7gM92R1yk1ScjMZY",
							"status":      "COMPLETED",
							"location_id": "L88917AVBK2S5",
						},
					},
				},
				NextPage: "bk9iU0RGNGdGZ0VxZ1pQ",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read payments last page omits the records key entirely",
			Input: common.ReadParams{
				ObjectName: "payments",
				Fields:     connectors.Fields("id", "status"),
				NextPage:   "bk9iU0RGNGdGZ0VxZ1pQ",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/payments"),
				Then:  mockserver.Response(http.StatusOK, []byte(`{}`)),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read customers ignores Since because the list endpoint has no time filter",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "given_name"),
				Since:      time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/customers"),
					mockcond.QueryParamsMissing("begin_time", "end_time"),
				},
				Then: mockserver.Response(http.StatusOK, customersLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "V9PD6FN1KMSQ8MEEYQHK0X1AC4",
						Fields: map[string]any{
							"id":         "V9PD6FN1KMSQ8MEEYQHK0X1AC4",
							"given_name": "Wilbur",
						},
						Raw: map[string]any{
							"id":         "V9PD6FN1KMSQ8MEEYQHK0X1AC4",
							"given_name": "Wilbur",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read payouts supports incremental begin_time and end_time",
			Input: common.ReadParams{
				ObjectName: "payouts",
				Fields:     connectors.Fields("id", "status"),
				Since:      time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/payouts"),
					mockcond.QueryParam("begin_time", "2022-03-01T00:00:00Z"),
					mockcond.QueryParam("end_time", "2022-03-31T00:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, payoutsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "po_b831a266-5fdf-4ea3-b1a0-bfff5c6cb0c9",
						Fields: map[string]any{
							"id":     "po_b831a266-5fdf-4ea3-b1a0-bfff5c6cb0c9",
							"status": "PAID",
						},
						Raw: map[string]any{
							"id":     "po_b831a266-5fdf-4ea3-b1a0-bfff5c6cb0c9",
							"status": "PAID",
							"type":   "BATCH",
						},
					},
				},
				NextPage: "TbfwAxZHHFvtVbpDgWd9F",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read merchants uses the singular merchant key and sends no limit",
			Input: common.ReadParams{
				ObjectName: "merchants",
				Fields:     connectors.Fields("id", "business_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/merchants"),
					mockcond.QueryParamsMissing("limit"),
				},
				Then: mockserver.Response(http.StatusOK, merchantsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "3MYCJG5GVYQ8Q",
						Fields: map[string]any{
							"id":            "3MYCJG5GVYQ8Q",
							"business_name": "Apple A Day",
						},
						Raw: map[string]any{
							"id":            "3MYCJG5GVYQ8Q",
							"business_name": "Apple A Day",
							"status":        "ACTIVE",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read gift_cards sends limit and reads the gift_cards array",
			Input: common.ReadParams{
				ObjectName: "gift_cards",
				Fields:     connectors.Fields("id", "state"),
				PageSize:   30,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/gift-cards"),
					mockcond.QueryParam("limit", "30"),
				},
				Then: mockserver.Response(http.StatusOK, giftCardsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "gftc:6cbacbb64cf54e2ca9f573d619038059",
						Fields: map[string]any{
							"id":    "gftc:6cbacbb64cf54e2ca9f573d619038059",
							"state": "ACTIVE",
						},
						Raw: map[string]any{
							"id":    "gftc:6cbacbb64cf54e2ca9f573d619038059",
							"state": "ACTIVE",
							"type":  "DIGITAL",
						},
					},
				},
				NextPage: "JbFx7dkqZ8oN1pXs",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read cards paginates by cursor and sends no limit",
			Input: common.ReadParams{
				ObjectName: "cards",
				Fields:     connectors.Fields("id", "last_4"),
				NextPage:   "JbFx7dkqZ8aBcDeF",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/cards"),
					mockcond.QueryParam("cursor", "JbFx7dkqZ8aBcDeF"),
					mockcond.QueryParamsMissing("limit"),
				},
				Then: mockserver.Response(http.StatusOK, cardsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "ccof:uIbfRdHyiP99HKi1aD9jE7tpAj6w",
						Fields: map[string]any{
							"id":     "ccof:uIbfRdHyiP99HKi1aD9jE7tpAj6w",
							"last_4": "1111",
						},
						Raw: map[string]any{
							"id":         "ccof:uIbfRdHyiP99HKi1aD9jE7tpAj6w",
							"last_4":     "1111",
							"card_brand": "VISA",
						},
					},
				},
				NextPage: "JbFx7dkqZ8aBcDeF",
				Done:     false,
			},
			ExpectedErrs: nil,
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
