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
	itemResponse := testutils.DataFromFile(t, "item-read.json")
	accountEmptyResponse := testutils.DataFromFile(t, "account-read-empty.json")
	errorResponse := testutils.DataFromFile(t, "error-bad-request.json")

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
						Fields: buildFieldMetadata(map[string]string{
							"AccountSubType":     "string",
							"AccountType":        "string",
							"Active":             "boolean",
							"Classification":     "string",
							"domain":             "string",
							"sparse":             "boolean",
							"FullyQualifiedName": "string",
							"Name":               "string",
						}),
						FieldsMap: nil,
					},
					"customer": {
						DisplayName: "Customer",
						Fields: buildFieldMetadata(map[string]string{
							"domain":                  "string",
							"FamilyName":              "string",
							"DisplayName":             "string",
							"PreferredDeliveryMethod": "string",
							"GivenName":               "string",
							"FullyQualifiedName":      "string",
							"BillWithParent":          "boolean",
							"Job":                     "boolean",
						}),
						FieldsMap: nil,
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe Item object with metadata",
			Input: []string{"item"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Item STARTPOSITION 0 MAXRESULTS 1"),
				Then:  mockserver.Response(http.StatusOK, itemResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"item": {
						DisplayName: "Item",
						Fields: buildFieldMetadata(map[string]string{
							"Name":   "string",
							"Type":   "string",
							"Active": "boolean",
							"domain": "string",
							"sparse": "boolean",
							"Level":  "string",
						}),
						FieldsMap: nil,
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Metadata fetch with empty results returns error",
			Input: []string{"account"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 0 MAXRESULTS 1"),
				Then:  mockserver.Response(http.StatusOK, accountEmptyResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrMissingExpectedValues},
		},
		{
			Name:  "Metadata fetch with error response",
			Input: []string{"account"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrBadRequest},
		},
		{
			Name:  "Successfully describe multiple objects including Item",
			Input: []string{"account", "customer", "item"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, accountResponse),
				}, {
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}, {
					If:   mockcond.QueryParam("query", "SELECT * FROM Item STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, itemResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"account": {
						DisplayName: "Account",
						Fields: buildFieldMetadata(map[string]string{
							"AccountSubType":     "string",
							"AccountType":        "string",
							"Active":             "boolean",
							"Classification":     "string",
							"domain":             "string",
							"sparse":             "boolean",
							"FullyQualifiedName": "string",
							"Name":               "string",
						}),
						FieldsMap: nil,
					},
					"customer": {
						DisplayName: "Customer",
						Fields: buildFieldMetadata(map[string]string{
							"domain":                  "string",
							"FamilyName":              "string",
							"DisplayName":             "string",
							"PreferredDeliveryMethod": "string",
							"GivenName":               "string",
							"FullyQualifiedName":      "string",
							"BillWithParent":          "boolean",
							"Job":                     "boolean",
						}),
						FieldsMap: nil,
					},
					"item": {
						DisplayName: "Item",
						Fields: buildFieldMetadata(map[string]string{
							"Name":   "string",
							"Type":   "string",
							"Active": "boolean",
							"domain": "string",
							"sparse": "boolean",
							"Level":  "string",
						}),
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

func buildFieldMetadata(fields map[string]string) map[string]common.FieldMetadata {
	result := make(map[string]common.FieldMetadata)
	for name, typ := range fields {
		result[name] = common.FieldMetadata{
			DisplayName:  name,
			ValueType:    common.ValueType(typ),
			ProviderType: "",
			Values:       nil,
		}
	}

	return result
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
