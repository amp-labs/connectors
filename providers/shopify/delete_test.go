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

func TestDelete(t *testing.T) {
	t.Parallel()

	responseCustomerDelete := testutils.DataFromFile(t, "delete/response-customer-delete.json")
	responseAddressDelete := testutils.DataFromFile(t, "delete/response-address-delete.json")

	requestCustomerDelete := testutils.DataFromFile(t, "delete/request-customer-delete.json")
	requestAddressDelete := testutils.DataFromFile(t, "delete/request-address-delete.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object name must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete record ID must be included",
			Input:        common.DeleteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "Successful customer delete",
			Input: common.DeleteParams{
				ObjectName: "customers",
				RecordId:   "gid://shopify/Customer/1073340122",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/api/2025-01/graphql.json"),
					mockcond.Body(string(requestCustomerDelete)),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerDelete),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful customer address delete",
			Input: common.DeleteParams{
				ObjectName: "customerAddresses",
				// Format: "customerId|addressId"
				RecordId: "gid://shopify/Customer/1018520244|gid://shopify/MailingAddress/1053318591",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/api/2025-01/graphql.json"),
					mockcond.Body(string(requestAddressDelete)),
				},
				Then: mockserver.Response(http.StatusOK, responseAddressDelete),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
