package paddle

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

	responseCustomersFirst := testutils.DataFromFile(t, "customers-first-page.json")
	responseCustomersSecond := testutils.DataFromFile(t, "customers-second-page.json")
	responseProducts := testutils.DataFromFile(t, "products.json")

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
				Then:  mockserver.Response(http.StatusOK, responseCustomersFirst),
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
				Then:  mockserver.Response(http.StatusOK, responseCustomersSecond),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read of products with chosen fields",
			Input: common.ReadParams{ObjectName: "products", Fields: connectors.Fields("id", "name", "status", "description")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/products"),
				Then:  mockserver.Response(http.StatusOK, responseProducts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":          "pro_01h1vjes1y163xfj1rh1tkfb65",
						"name":        "Analytics addon",
						"status":      "active",
						"description": "Unlock advanced insights.",
					},
					Raw: map[string]any{
						"id":           "pro_01h1vjes1y163xfj1rh1tkfb65",
						"name":         "Analytics addon",
						"tax_category": "standard",
						"type":         "standard",
						"description":  "Unlock advanced insights.",
						"image_url":    "https://paddle.s3.amazonaws.com/user/165798/97dRpA6SXzcE6ekK9CAr_analytics.png",
						"custom_data":  nil,
						"status":       "active",
						"import_meta":  nil,
						"created_at":   "2023-06-01T13:30:50.302Z",
						"updated_at":   "2024-04-05T15:47:17.163Z",
					},
				}},
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
