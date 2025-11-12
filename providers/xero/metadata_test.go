package xero

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	accountResponse := testutils.DataFromFile(t, "accounts.json")
	contactsResponse := testutils.DataFromFile(t, "contacts.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"accounts", "contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api.xro/2.0/Accounts"),
					Then: mockserver.Response(http.StatusOK, accountResponse),
				}, {
					If:   mockcond.Path("/api.xro/2.0/Contacts"),
					Then: mockserver.Response(http.StatusOK, contactsResponse),
				}},
			}.Server(),

			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"accounts": {
						DisplayName: "Accounts",
						Fields: map[string]common.FieldMetadata{
							"AccountID": {
								DisplayName:  "AccountID",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"Code": {
								DisplayName:  "Code",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"EnablePaymentsToAccount": {
								DisplayName:  "EnablePaymentsToAccount",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"Name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"TaxType": {
								DisplayName:  "TaxType",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"Type": {
								DisplayName:  "Type",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"BankAccountNumber": {
								DisplayName:  "BankAccountNumber",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"BankAccountType": {
								DisplayName:  "BankAccountType",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"CurrencyCode": {
								DisplayName:  "CurrencyCode",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"ContactID": {
								DisplayName:  "ContactID",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"ContactStatus": {
								DisplayName:  "ContactStatus",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"Name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"FirstName": {
								DisplayName:  "FirstName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"LastName": {
								DisplayName:  "LastName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"CompanyNumber": {
								DisplayName:  "CompanyNumber",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"EmailAddress": {
								DisplayName:  "EmailAddress",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"BankAccountDetails": {
								DisplayName:  "BankAccountDetails",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"TaxNumber": {
								DisplayName:  "TaxNumber",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"AccountsReceivableTaxType": {
								DisplayName:  "AccountsReceivableTaxType",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"AccountsPayableTaxType": {
								DisplayName:  "AccountsPayableTaxType",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"Addresses": {
								DisplayName:  "Addresses",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"Phones": {
								DisplayName:  "Phones",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"UpdatedDateUTC": {
								DisplayName:  "UpdatedDateUTC",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"IsSupplier": {
								DisplayName:  "IsSupplier",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"IsCustomer": {
								DisplayName:  "IsCustomer",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"DefaultCurrency": {
								DisplayName:  "DefaultCurrency",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
