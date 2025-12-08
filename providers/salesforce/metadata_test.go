package salesforce

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseOrgMeta := testutils.DataFromFile(t, "metadata-organization-sampled.json")
	responseCustomObjMeta := testutils.DataFromFile(t, "metadata/custom-object-with-custom-fields.json")

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
					"url":"/services/data/v60.0/sobjects/Organization/describe"
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
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
								ReadOnly:     false,
							},
							"preferencesconsentmanagementenabled": {
								DisplayName:  "ConsentManagementEnabled",
								ValueType:    "boolean",
								ProviderType: "boolean",
								ReadOnly:     false,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
								Values:       nil,
							},
							"latitude": {
								DisplayName:  "Latitude",
								ValueType:    "float",
								ProviderType: "double",
								ReadOnly:     false,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							// Nested field: Latitude is a component of the Address compound field.
							// It appears both as a flat field ("latitude") and as a nested field.
							"$['address']['latitude']": {
								DisplayName:  "Latitude",
								ValueType:    "float",
								ProviderType: "double",
								ReadOnly:     false,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"monthlypageviewsused": {
								DisplayName:  "Monthly Page Views Used",
								ValueType:    "int",
								ProviderType: "int",
								ReadOnly:     true,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"systemmodstamp": {
								DisplayName:  "System Modstamp",
								ValueType:    "datetime",
								ProviderType: "datetime",
								ReadOnly:     true,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
								Values:       nil,
							},
							"defaultaccountaccess": {
								DisplayName:  "Default Account Access",
								ValueType:    "singleSelect",
								ProviderType: "picklist",
								ReadOnly:     true,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
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
								ReadOnly:     false,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
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
		{
			Name:  "Custom and required fields",
			Input: []string{"TestObject15__c"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{"allOrNone":false,"compositeRequest":[{
					"referenceId":"TestObject15__c",
					"method":"GET",
					"url":"/services/data/v60.0/sobjects/TestObject15__c/describe"
				}]}`),
				Then: mockserver.Response(http.StatusOK, responseCustomObjMeta),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"testobject15__c": {
						DisplayName: "Test Object 15",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Record ID",
								ValueType:    "other",
								ProviderType: "id",
								ReadOnly:     true,
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
							},
							"interests__c": {
								DisplayName:  "Interests",
								ValueType:    "multiSelect",
								ProviderType: "multipicklist",
								ReadOnly:     false,
								IsCustom:     goutils.Pointer(true),
								IsRequired:   goutils.Pointer(true),
								Values: []common.FieldValue{{
									Value:        "art",
									DisplayValue: "art",
								}, {
									Value:        "swimming",
									DisplayValue: "swimming",
								}, {
									Value:        "travel",
									DisplayValue: "travel",
								}},
							},
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

func TestListObjectMetadataPardot(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Successfully describe one object with metadata",
			Input:      []string{"EmAiLs"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"emails": {
						DisplayName: "Emails",
						Fields: map[string]common.FieldMetadata{
							"htmlMessage": {
								DisplayName:  "HTML Message",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"sentAt": {
								DisplayName:  "Sent At",
								ValueType:    "datetime",
								ProviderType: "datetime",
								ReadOnly:     true,
								Values:       nil,
							},
							"type": {
								DisplayName:  "Type",
								ValueType:    "singleSelect",
								ProviderType: "enum",
								ReadOnly:     true,
								Values: []common.FieldValue{{
									Value:        "html",
									DisplayValue: "HTML",
								}, {
									Value:        "text",
									DisplayValue: "Text",
								}, {
									Value:        "htmlAndText",
									DisplayValue: "HTML and Text",
								}},
							},
						},
						FieldsMap: map[string]string{
							"sentAt":          "Sent At",
							"subject":         "Subject",
							"textMessage":     "Text Message",
							"trackerDomainId": "Tracker Domain ID",
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
				return constructTestConnectorAccountEngagement(tt.Server.URL)
			})
		})
	}
}
