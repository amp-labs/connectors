package quickbooks

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
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
						Fields: common.FieldsMetadata{
							"AccountSubType":     {DisplayName: "AccountSubType", ValueType: "string"},
							"AccountType":        {DisplayName: "AccountType", ValueType: "string"},
							"Active":             {DisplayName: "Active", ValueType: "boolean"},
							"Classification":     {DisplayName: "Classification", ValueType: "string"},
							"domain":             {DisplayName: "domain", ValueType: "string"},
							"sparse":             {DisplayName: "sparse", ValueType: "boolean"},
							"FullyQualifiedName": {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"Name":               {DisplayName: "Name", ValueType: "string"},
						},
					},
					"customer": {
						DisplayName: "Customer",
						Fields: common.FieldsMetadata{
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
						},
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
						Fields: common.FieldsMetadata{
							"Name":   {DisplayName: "Name", ValueType: "string"},
							"Type":   {DisplayName: "Type", ValueType: "string"},
							"Active": {DisplayName: "Active", ValueType: "boolean"},
							"domain": {DisplayName: "domain", ValueType: "string"},
							"sparse": {DisplayName: "sparse", ValueType: "boolean"},
							"Level":  {DisplayName: "Level", ValueType: "string"},
						},
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
						Fields: common.FieldsMetadata{
							"AccountSubType":     {DisplayName: "AccountSubType", ValueType: "string"},
							"AccountType":        {DisplayName: "AccountType", ValueType: "string"},
							"Active":             {DisplayName: "Active", ValueType: "boolean"},
							"Classification":     {DisplayName: "Classification", ValueType: "string"},
							"domain":             {DisplayName: "domain", ValueType: "string"},
							"sparse":             {DisplayName: "sparse", ValueType: "boolean"},
							"FullyQualifiedName": {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"Name":               {DisplayName: "Name", ValueType: "string"},
						},
					},
					"customer": {
						DisplayName: "Customer",
						Fields: common.FieldsMetadata{
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
						},
					},
					"item": {
						DisplayName: "Item",
						Fields: common.FieldsMetadata{
							"Name":   {DisplayName: "Name", ValueType: "string"},
							"Type":   {DisplayName: "Type", ValueType: "string"},
							"Active": {DisplayName: "Active", ValueType: "boolean"},
							"domain": {DisplayName: "domain", ValueType: "string"},
							"sparse": {DisplayName: "sparse", ValueType: "boolean"},
							"Level":  {DisplayName: "Level", ValueType: "string"},
						},
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
						Fields: common.FieldsMetadata{
							// Base fields from REST API response
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
							// Custom fields from GraphQL response
							"ProjectCode":  {DisplayName: "ProjectCode", ValueType: "string", ProviderType: "StringType", IsCustom: goutils.Pointer(true)},
							"Department":   {DisplayName: "Department", ValueType: "string", ProviderType: "StringType", IsCustom: goutils.Pointer(true)},
							"BudgetAmount": {DisplayName: "BudgetAmount", ValueType: "float", ProviderType: "NumberType", IsCustom: goutils.Pointer(true)},
							"StartDate":    {DisplayName: "StartDate", ValueType: "datetime", ProviderType: "DateType", IsCustom: goutils.Pointer(true)},
							"Status":       {DisplayName: "Status", ValueType: "singleSelect", ProviderType: "ListType", IsCustom: goutils.Pointer(true)},
						},
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
						Fields: common.FieldsMetadata{
							// Base fields from REST API response
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
							// Custom fields from GraphQL response
							"ProjectCode":  {DisplayName: "ProjectCode", ValueType: "string", ProviderType: "StringType", IsCustom: goutils.Pointer(true)},
							"Department":   {DisplayName: "Department", ValueType: "string", ProviderType: "StringType", IsCustom: goutils.Pointer(true)},
							"BudgetAmount": {DisplayName: "BudgetAmount", ValueType: "float", ProviderType: "NumberType", IsCustom: goutils.Pointer(true)},
							"StartDate":    {DisplayName: "StartDate", ValueType: "datetime", ProviderType: "DateType", IsCustom: goutils.Pointer(true)},
							"Status":       {DisplayName: "Status", ValueType: "singleSelect", ProviderType: "ListType", IsCustom: goutils.Pointer(true)},
						},
					},
					"account": {
						DisplayName: "Account",
						Fields: common.FieldsMetadata{
							"AccountSubType":     {DisplayName: "AccountSubType", ValueType: "string"},
							"AccountType":        {DisplayName: "AccountType", ValueType: "string"},
							"Active":             {DisplayName: "Active", ValueType: "boolean"},
							"Classification":     {DisplayName: "Classification", ValueType: "string"},
							"domain":             {DisplayName: "domain", ValueType: "string"},
							"sparse":             {DisplayName: "sparse", ValueType: "boolean"},
							"FullyQualifiedName": {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"Name":               {DisplayName: "Name", ValueType: "string"},
						},
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
						Fields: common.FieldsMetadata{
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
						},
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
						Fields: common.FieldsMetadata{
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
						},
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
						Fields: common.FieldsMetadata{
							"domain":                  {DisplayName: "domain", ValueType: "string"},
							"FamilyName":              {DisplayName: "FamilyName", ValueType: "string"},
							"DisplayName":             {DisplayName: "DisplayName", ValueType: "string"},
							"PreferredDeliveryMethod": {DisplayName: "PreferredDeliveryMethod", ValueType: "string"},
							"GivenName":               {DisplayName: "GivenName", ValueType: "string"},
							"FullyQualifiedName":      {DisplayName: "FullyQualifiedName", ValueType: "string"},
							"BillWithParent":          {DisplayName: "BillWithParent", ValueType: "boolean"},
							"Job":                     {DisplayName: "Job", ValueType: "boolean"},
						},
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

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))
	connector.graphQLBaseURL = serverURL + "/graphql"

	return connector, nil
}
