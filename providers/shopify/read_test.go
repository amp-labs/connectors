package shopify

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

func TestRead(t *testing.T) {
	t.Parallel()

	responseProducts := testutils.DataFromFile(t, "read/products.json")
	responseProductsWithNextPage := testutils.DataFromFile(t, "read/products-with-next-page.json")
	responseCustomers := testutils.DataFromFile(t, "read/customers.json")
	responseOrders := testutils.DataFromFile(t, "read/orders.json")
	responseErrorInvalidQuery := testutils.DataFromFile(t, "read/error-invalid-query.json")

	requestProducts := testutils.DataFromFile(t, "read/request/products.json")
	requestCustomers := testutils.DataFromFile(t, "read/request/customers.json")
	requestOrders := testutils.DataFromFile(t, "read/request/orders.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "products"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error response with invalid query",
			Input: common.ReadParams{ObjectName: "products", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseErrorInvalidQuery),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
			},
		},
		{
			Name:  "Successfully read products",
			Input: common.ReadParams{ObjectName: "products", Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestProducts)),
				},
				Then: mockserver.Response(http.StatusOK, responseProducts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "gid://shopify/Product/123456789",
							"title": "Test Product",
						},
						Raw: map[string]any{
							"id":     "gid://shopify/Product/123456789",
							"title":  "Test Product",
							"handle": "test-product",
							"status": "ACTIVE",
						},
					},
					{
						Fields: map[string]any{
							"id":    "gid://shopify/Product/987654321",
							"title": "Another Product",
						},
						Raw: map[string]any{
							"id":     "gid://shopify/Product/987654321",
							"title":  "Another Product",
							"handle": "another-product",
							"status": "DRAFT",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully read products with pagination",
			Input: common.ReadParams{ObjectName: "products", Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestProducts)),
				},
				Then: mockserver.Response(http.StatusOK, responseProductsWithNextPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "gid://shopify/Product/123456789",
							"title": "Test Product",
						},
						Raw: map[string]any{
							"id":     "gid://shopify/Product/123456789",
							"title":  "Test Product",
							"handle": "test-product",
							"status": "ACTIVE",
						},
					},
				},
				NextPage: "eyJsYXN0X2lkIjoxMjM0NTY3ODksImxhc3RfdmFsdWUiOiIyMDI0LTAxLTIwVDE0OjQ1OjAwWiJ9",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully read customers",
			Input: common.ReadParams{ObjectName: "customers", Fields: connectors.Fields("id", "firstname", "displayname")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestCustomers)),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "gid://shopify/Customer/111111111",
							"firstname":   "John",
							"displayname": "John Doe",
						},
						Raw: map[string]any{
							"id":          "gid://shopify/Customer/111111111",
							"firstName":   "John",
							"lastName":    "Doe",
							"displayName": "John Doe",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully read orders",
			Input: common.ReadParams{ObjectName: "orders", Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestOrders)),
				},
				Then: mockserver.Response(http.StatusOK, responseOrders),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":   "gid://shopify/Order/555555555",
							"name": "#1001",
						},
						Raw: map[string]any{
							"id":    "gid://shopify/Order/555555555",
							"name":  "#1001",
							"email": "customer@example.com",
						},
					},
				},
				NextPage: "",
				Done:     true,
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
