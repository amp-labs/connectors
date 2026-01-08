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
	graphQLResponse := testutils.DataFromFile(t, "custom-fields/graphql-response.json")
	graphQLEmptyResponse := testutils.DataFromFile(t, "custom-fields/graphql-empty.json")
	graphQLErrorResponse := testutils.DataFromFile(t, "custom-fields/graphql-error.json")

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
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"account": common.ErrMissingExpectedValues,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Metadata fetch with error response",
			Input: []string{"account"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"account": mockutils.ExpectedSubsetErrors{
						common.ErrCaller,
					},
				},
			},
			ExpectedErrs: nil,
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
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe Customer with custom fields",
			Input: []string{"customer"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/graphql"),
					},
					Then: mockserver.Response(http.StatusOK, graphQLResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customer": {
						DisplayName: "Customer",
						Fields: buildFieldMetadataWithCustomFields(map[string]string{
							"domain":                  "string",
							"FamilyName":              "string",
							"DisplayName":             "string",
							"PreferredDeliveryMethod": "string",
							"GivenName":               "string",
							"FullyQualifiedName":      "string",
							"BillWithParent":          "boolean",
							"Job":                     "boolean",
							"ProjectCode":             "string",
							"Department":              "string",
							"BudgetAmount":            "float",
							"StartDate":               "datetime",
							"Status":                  "singleSelect",
						}),
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe Customer and Account with custom fields (mixed objects)",
			Input: []string{"customer", "account"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}, {
					If:   mockcond.QueryParam("query", "SELECT * FROM Account STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, accountResponse),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/graphql"),
					},
					Then: mockserver.Response(http.StatusOK, graphQLResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customer": {
						DisplayName: "Customer",
						Fields: buildFieldMetadataWithCustomFields(map[string]string{
							"domain":                  "string",
							"FamilyName":              "string",
							"DisplayName":             "string",
							"PreferredDeliveryMethod": "string",
							"GivenName":               "string",
							"FullyQualifiedName":      "string",
							"BillWithParent":          "boolean",
							"Job":                     "boolean",
							"ProjectCode":             "string",
							"Department":              "string",
							"BudgetAmount":            "float",
							"StartDate":               "datetime",
							"Status":                  "singleSelect",
						}),
					},
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
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "GraphQL failure gracefully degrades to base metadata",
			Input: []string{"customer"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/graphql"),
					},
					Then: mockserver.Response(http.StatusInternalServerError),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
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
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Empty custom fields response returns base metadata only",
			Input: []string{"customer"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/graphql"),
					},
					Then: mockserver.Response(http.StatusOK, graphQLEmptyResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
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
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "GraphQL errors in response return error",
			Input: []string{"customer"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("query", "SELECT * FROM Customer STARTPOSITION 0 MAXRESULTS 1"),
					Then: mockserver.Response(http.StatusOK, customerResponse),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/graphql"),
					},
					Then: mockserver.Response(http.StatusOK, graphQLErrorResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
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
	return buildFieldMetadataWithProviderTypes(fields, nil)
}

func buildFieldMetadataWithCustomFields(fields map[string]string) map[string]common.FieldMetadata {
	providerTypes := map[string]string{
		"ProjectCode":  "StringType",
		"Department":   "StringType",
		"BudgetAmount": "NumberType",
		"StartDate":    "DateType",
		"Status":       "ListType",
	}
	return buildFieldMetadataWithProviderTypes(fields, providerTypes)
}

func buildFieldMetadataWithProviderTypes(fields map[string]string, providerTypes map[string]string) map[string]common.FieldMetadata {
	result := make(map[string]common.FieldMetadata)
	for name, typ := range fields {
		providerType := ""
		if providerTypes != nil {
			providerType = providerTypes[name]
		}
		result[name] = common.FieldMetadata{
			DisplayName:  name,
			ValueType:    common.ValueType(typ),
			ProviderType: providerType,
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

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))
	connector.graphQLBaseURL = serverURL + "/graphql"

	return connector, nil
}
