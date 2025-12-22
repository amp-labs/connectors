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
	responseAddressCreate := testutils.DataFromFile(t, "write/response-address-create.json")
	responseAddressUpdate := testutils.DataFromFile(t, "write/response-address-update.json")
	responseDefaultAddressUpdate := testutils.DataFromFile(t, "write/response-default-address-update.json")

	requestCustomerCreate := testutils.DataFromFile(t, "write/request-customer-create.json")
	requestCustomerUpdate := testutils.DataFromFile(t, "write/request-customer-update.json")
	requestAddressCreate := testutils.DataFromFile(t, "write/request-address-create.json")
	requestAddressUpdate := testutils.DataFromFile(t, "write/request-address-update.json")
	requestDefaultAddressUpdate := testutils.DataFromFile(t, "write/request-default-address-update.json")

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
					mockcond.Path(testApiPath),
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
					mockcond.Path(testApiPath),
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
			Name: "Successful customer address create",
			Input: common.WriteParams{
				ObjectName: "customerAddresses",
				RecordData: map[string]any{
					"customerId": "gid://shopify/Customer/1018520244",
					"address": map[string]any{
						"address1":     "123 Main St",
						"city":         "Ottawa",
						"countryCode":  "CA",
						"firstName":    "Steve",
						"lastName":     "Lastname",
						"phone":        "+16469999999",
						"provinceCode": "ON",
						"zip":          "A1A 4A1",
					},
					"setAsDefault": true,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestAddressCreate)),
				},
				Then: mockserver.Response(http.StatusOK, responseAddressCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/MailingAddress/1053318591",
				Data: map[string]any{
					"id": "gid://shopify/MailingAddress/1053318591",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful customer address update",
			Input: common.WriteParams{
				ObjectName: "customerAddresses",
				RecordId:   "gid://shopify/MailingAddress/1053318591",
				RecordData: map[string]any{
					"customerId": "gid://shopify/Customer/1018520244",
					"address": map[string]any{
						"address1":     "456 Updated St",
						"city":         "Toronto",
						"countryCode":  "CA",
						"firstName":    "Steve",
						"lastName":     "Updated",
						"provinceCode": "ON",
						"zip":          "M5V 1J1",
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestAddressUpdate)),
				},
				Then: mockserver.Response(http.StatusOK, responseAddressUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/MailingAddress/1053318591",
				Data: map[string]any{
					"id": "gid://shopify/MailingAddress/1053318591",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful customer default address update",
			Input: common.WriteParams{
				ObjectName: "customerDefaultAddress",
				RecordData: map[string]any{
					"customerId": "gid://shopify/Customer/624407574",
					"addressId":  "gid://shopify/MailingAddress/1053318591?model_name=CustomerAddress",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testApiPath),
					mockcond.Body(string(requestDefaultAddressUpdate)),
				},
				Then: mockserver.Response(http.StatusOK, responseDefaultAddressUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "gid://shopify/MailingAddress/1053318591?model_name=CustomerAddress",
				Data: map[string]any{
					"id": "gid://shopify/MailingAddress/1053318591?model_name=CustomerAddress",
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
