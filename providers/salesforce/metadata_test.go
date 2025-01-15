package salesforce

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseOrgMeta := testutils.DataFromFile(t, "metadata-organization-sampled.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Mime response header expected for successful response",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Always: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			ExpectedErrs: []error{common.ErrNotJSON},
		},
		{
			Name:  "Successfully describe one object with metadata",
			Input: []string{"Organization"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{"allOrNone":false,"compositeRequest":[{
					"referenceId":"Organization",
					"method":"GET",
					"url":"/services/data/v59.0/sobjects/Organization/describe"
				}]}`),
				Then: mockserver.Response(http.StatusOK, responseOrgMeta),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"organization": {
						DisplayName: "Organization",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     true,
								Values:       nil,
							},
							"preferencesconsentmanagementenabled": {
								DisplayName:  "ConsentManagementEnabled",
								ValueType:    "boolean",
								ProviderType: "boolean",
								ReadOnly:     true,
								Values:       nil,
							},
							"latitude": {
								DisplayName:  "Latitude",
								ValueType:    "float",
								ProviderType: "double",
								ReadOnly:     true,
								Values:       nil,
							},
							"monthlypageviewsused": {
								DisplayName:  "Monthly Page Views Used",
								ValueType:    "int",
								ProviderType: "int",
								ReadOnly:     true,
								Values:       nil,
							},
							"systemmodstamp": {
								DisplayName:  "System Modstamp",
								ValueType:    "datetime",
								ProviderType: "datetime",
								ReadOnly:     true,
								Values:       nil,
							},
							"defaultaccountaccess": {
								DisplayName:  "Default Account Access",
								ValueType:    "singleSelect",
								ProviderType: "picklist",
								ReadOnly:     true,
								Values: []common.FieldValue{{
									Value:        "None",
									DisplayValue: "Private",
								}, {
									Value:        "Read",
									DisplayValue: "Read Only",
								}, {
									Value:        "Edit",
									DisplayValue: "Read/Write",
								}, {
									Value:        "ControlledByLeadOrContact",
									DisplayValue: "Controlled By Lead Or Contact",
								}, {
									Value:        "ControlledByCampaign",
									DisplayValue: "Controlled By Campaign",
								}},
							},
							"phone": {
								DisplayName:  "Phone",
								ValueType:    "other",
								ProviderType: "phone",
								ReadOnly:     true,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"defaultaccountaccess":                "Default Account Access",
							"latitude":                            "Latitude",
							"monthlypageviewsused":                "Monthly Page Views Used",
							"name":                                "Name",
							"phone":                               "Phone",
							"preferencesconsentmanagementenabled": "ConsentManagementEnabled",
							"systemmodstamp":                      "System Modstamp",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
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
