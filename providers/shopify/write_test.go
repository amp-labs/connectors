package shopify

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseCustomerCreate := testutils.DataFromFile(t, "write/response-customer-create.json")
	responseCustomerCreateError := testutils.DataFromFile(t, "write/response-customer-create-error.json")
	responseCustomerUpdate := testutils.DataFromFile(t, "write/response-customer-update.json")
	responseProductCreate := testutils.DataFromFile(t, "write/response-product-create.json")
	responseProductUpdate := testutils.DataFromFile(t, "write/response-product-update.json")

	requestCustomerCreate := testutils.DataFromFile(t, "write/request-customer-create.json")
	requestCustomerUpdate := testutils.DataFromFile(t, "write/request-customer-update.json")
	requestProductCreate := testutils.DataFromFile(t, "write/request-product-create.json")
	requestProductUpdate := testutils.DataFromFile(t, "write/request-product-update.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Successful customer create",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{
					"email":     "steve.lastnameson@example.com",
					"firstName": "Steve",
					"phone":     "+16465555555",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/api/2025-10/graphql.json"),
					mockcond.Body(string(requestCustomerCreate)),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/Customer/1073340122",
				Data: map[string]any{
					"id": "gid://shopify/Customer/1073340122",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Customer create with validation error",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseCustomerCreateError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Customer must have a name, phone number or email address"),
			},
		},
		{
			Name: "Successful customer update",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "gid://shopify/Customer/1018520244",
				RecordData: map[string]any{
					"firstName": "Tobi",
					"lastName":  "Lutke",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/api/2025-10/graphql.json"),
					mockcond.Body(string(requestCustomerUpdate)),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/Customer/1018520244",
				Data: map[string]any{
					"id": "gid://shopify/Customer/1018520244",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful product create",
			Input: common.WriteParams{
				ObjectName: "products",
				RecordData: map[string]any{
					"title":       "Cool socks",
					"productType": "Apparel",
					"vendor":      "TestVendor",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/api/2025-10/graphql.json"),
					mockcond.Body(string(requestProductCreate)),
				},
				Then: mockserver.Response(http.StatusOK, responseProductCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/Product/1072482054",
				Data: map[string]any{
					"id": "gid://shopify/Product/1072482054",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful product update",
			Input: common.WriteParams{
				ObjectName: "products",
				RecordId:   "gid://shopify/Product/1072482054",
				RecordData: map[string]any{
					"title":  "Updated Cool socks",
					"status": "ACTIVE",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/api/2025-10/graphql.json"),
					mockcond.Body(string(requestProductUpdate)),
				},
				Then: mockserver.Response(http.StatusOK, responseProductUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/Product/1072482054",
				Data: map[string]any{
					"id": "gid://shopify/Product/1072482054",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
