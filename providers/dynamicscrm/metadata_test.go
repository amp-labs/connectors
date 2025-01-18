package dynamicscrm

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

	responseContactsEntityDefinition := testutils.DataFromFile(t, "metadata/contacts/entity-definition.json")
	// Attributes file is a shorter form of real Microsoft server response.
	responseContactsAttributes := testutils.DataFromFile(t, "metadata/contacts/attributes.json")
	responseContactsOptionsPicklists := testutils.DataFromFile(t, "metadata/contacts/options-picklists.json")
	responseContactsOptionsStates := testutils.DataFromFile(t, "metadata/contacts/options-states.json")
	responseContactsOptionsStatuses := testutils.DataFromFile(t, "metadata/contacts/options-statuses.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Failure to return attributes for an object",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": mockutils.ExpectedSubsetErrors{
						ErrFetchAttributes,
						ErrObjectNotFound,
					},
				},
			},
		},
		{
			Name:  "Provider returns empty list of attributes",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')"),
					Then: mockserver.Response(http.StatusOK, responseContactsEntityDefinition),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes"),
					Then: mockserver.ResponseString(http.StatusOK, `{"value":[]}`), // no object attributes
				}},
				Default: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"contacts": mockutils.ExpectedSubsetErrors{
						ErrObjectMissingAttributes,
					},
				},
			},
		},
		{
			Name:  "Failure to fetch PicklistType, StatusType and StateType enumeration options",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')"),
					Then: mockserver.Response(http.StatusOK, responseContactsEntityDefinition),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes"),
					Then: mockserver.Response(http.StatusOK, responseContactsAttributes),
				}},
				Default: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"contacts": mockutils.ExpectedSubsetErrors{
						ErrFetchAttributesPicklists,
						ErrFetchAttributesStatuses,
						ErrFetchAttributesStates,
						common.ErrEmptyJSONHTTPResponse,
					},
				},
			},
		},
		{
			Name:  "Successfully collect and combine metadata for an object",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')"),
					Then: mockserver.Response(http.StatusOK, responseContactsEntityDefinition),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes"),
					Then: mockserver.Response(http.StatusOK, responseContactsAttributes),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes/Microsoft.Dynamics.CRM.PicklistAttributeMetadata"), // nolint:lll
					Then: mockserver.Response(http.StatusOK, responseContactsOptionsPicklists),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes/Microsoft.Dynamics.CRM.StatusAttributeMetadata"), // nolint:lll
					Then: mockserver.Response(http.StatusOK, responseContactsOptionsStatuses),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes/Microsoft.Dynamics.CRM.StateAttributeMetadata"), // nolint:lll
					Then: mockserver.Response(http.StatusOK, responseContactsOptionsStates),
				}},
				Default: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"adx_identity_newpassword": {
								DisplayName:  "New Password Input",
								ValueType:    "string",
								ProviderType: "StringType",
								ReadOnly:     false,
								Values:       nil,
							},
							"adx_publicprofilecopy": {
								DisplayName:  "Public Profile Copy",
								ValueType:    "string",
								ProviderType: "MemoType",
								ReadOnly:     false,
								Values:       nil,
							},
							"merged": {
								DisplayName:  "Merged",
								ValueType:    "boolean",
								ProviderType: "BooleanType",
								ReadOnly:     true,
								Values:       nil,
							},
							"versionnumber": {
								DisplayName:  "Version Number",
								ValueType:    "int",
								ProviderType: "BigIntType",
								ReadOnly:     true,
								Values:       nil,
							},
							"importsequencenumber": {
								DisplayName:  "Import Sequence Number",
								ValueType:    "int",
								ProviderType: "IntegerType",
								ReadOnly:     false,
								Values:       nil,
							},
							"exchangerate": {
								DisplayName:  "Exchange Rate",
								ValueType:    "float",
								ProviderType: "DecimalType",
								ReadOnly:     true,
								Values:       nil,
							},
							"annualincome": {
								DisplayName:  "Annual Income",
								ValueType:    "float",
								ProviderType: "MoneyType",
								ReadOnly:     false,
								Values:       nil,
							},
							"birthdate": {
								DisplayName:  "Birthday",
								ValueType:    "date",
								ProviderType: "DateTimeType",
								ReadOnly:     false,
								Values:       nil,
							},
							"createdon": {
								DisplayName:  "Created On",
								ValueType:    "datetime",
								ProviderType: "DateTimeType",
								ReadOnly:     true,
								Values:       nil,
							},
							"statecode": {
								DisplayName:  "Status",
								ValueType:    "singleSelect",
								ProviderType: "StateType",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "0",
									DisplayValue: "Active",
								}, {
									Value:        "1",
									DisplayValue: "Inactive",
								}},
							},
							"statuscode": {
								DisplayName:  "Status Reason",
								ValueType:    "singleSelect",
								ProviderType: "StatusType",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "1",
									DisplayValue: "Active",
								}, {
									Value:        "2",
									DisplayValue: "Inactive",
								}},
							},
							"gendercode": {
								DisplayName:  "Gender",
								ValueType:    "singleSelect",
								ProviderType: "PicklistType",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "1",
									DisplayValue: "Male",
								}, {
									Value:        "2",
									DisplayValue: "Female",
								}},
							},
							"familystatuscode": {
								DisplayName:  "Marital Status",
								ValueType:    "singleSelect",
								ProviderType: "PicklistType",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "1",
									DisplayValue: "Single",
								}, {
									Value:        "2",
									DisplayValue: "Married",
								}, {
									Value:        "3",
									DisplayValue: "Divorced",
								}, {
									Value:        "4",
									DisplayValue: "Widowed",
								}},
							},
							"educationcode": {
								DisplayName:  "Education",
								ValueType:    "singleSelect",
								ProviderType: "PicklistType",
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "1",
									DisplayValue: "Default Value",
								}},
							},
							"leadsourcecodename": {
								DisplayName:  "LeadSourceCodeName",
								ValueType:    "other",
								ProviderType: "VirtualType",
								ReadOnly:     true,
								Values:       nil,
							},
							"_accountid_value": {
								DisplayName:  "Account",
								ValueType:    "other",
								ProviderType: "LookupType",
								ReadOnly:     true,
								Values:       nil,
							},
							"_createdby_value": {
								DisplayName:  "Created By",
								ValueType:    "other",
								ProviderType: "LookupType",
								ReadOnly:     true,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							// nice display names
							"adx_publicprofilecopy":    "Public Profile Copy",
							"adx_identity_newpassword": "New Password Input",
							// schema name was used for display
							"leadsourcecodename": "LeadSourceCodeName",
							// underscore prefixed fields
							"_accountid_value": "Account",
							"_createdby_value": "Created By",
						},
					},
				},
				Errors: nil,
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
