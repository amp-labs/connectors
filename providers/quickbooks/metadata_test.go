package quickbooks

import (
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

	accountResponse := testutils.DataFromFile(t, "account-read.json")
	customerResponse := testutils.DataFromFile(t, "customer-read.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"account", "customer"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, accountResponse),
				}, {
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"account": {
						DisplayName: "Account",
						Fields: map[string]common.FieldMetadata{
							"AccountSubType": {
								DisplayName:  "AccountSubType",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"AccountType": {
								DisplayName:  "AccountType",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"Active": {
								DisplayName:  "Active",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"Classification": {
								DisplayName:  "Classification",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"domain": {
								DisplayName:  "domain",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"sparse": {
								DisplayName:  "sparse",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"FullyQualifiedName": {
								DisplayName:  "FullyQualifiedName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"Name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
					"customer": {
						DisplayName: "Customer",
						Fields: map[string]common.FieldMetadata{
							"domain": {
								DisplayName:  "domain",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"FamilyName": {
								DisplayName:  "FamilyName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"DisplayName": {
								DisplayName:  "DisplayName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"PreferredDeliveryMethod": {
								DisplayName:  "PreferredDeliveryMethod",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"GivenName": {
								DisplayName:  "GivenName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"FullyQualifiedName": {
								DisplayName:  "FullyQualifiedName",
								ValueType:    "string",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"BillWithParent": {
								DisplayName:  "BillWithParent",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"Job": {
								DisplayName:  "Job",
								ValueType:    "boolean",
								ProviderType: "",
								ReadOnly:     false,
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
		Metadata: map[string]string{
			"realmID": "123456789",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
