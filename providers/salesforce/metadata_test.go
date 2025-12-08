package salesforce

import (
	"errors"
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
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
							},
							"preferencesconsentmanagementenabled": {
								DisplayName:  "ConsentManagementEnabled",
								ValueType:    "boolean",
								ProviderType: "boolean",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
								Values:       nil,
							},
							"latitude": {
								DisplayName:  "Latitude",
								ValueType:    "float",
								ProviderType: "double",
								ReadOnly:     goutils.Pointer(false),
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
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"systemmodstamp": {
								DisplayName:  "System Modstamp",
								ValueType:    "datetime",
								ProviderType: "datetime",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
								Values:       nil,
							},
							"defaultaccountaccess": {
								DisplayName:  "Default Account Access",
								ValueType:    "singleSelect",
								ProviderType: "picklist",
								ReadOnly:     goutils.Pointer(true),
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
								ReadOnly:     goutils.Pointer(false),
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
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
							},
							"interests__c": {
								DisplayName:  "Interests",
								ValueType:    "multiSelect",
								ProviderType: "multipicklist",
								ReadOnly:     goutils.Pointer(false),
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
								Values:       nil,
							},
							"sentAt": {
								DisplayName:  "Sent At",
								ValueType:    "datetime",
								ProviderType: "datetime",
								ReadOnly:     goutils.Pointer(true),
								Values:       nil,
							},
							"type": {
								DisplayName:  "Type",
								ValueType:    "singleSelect",
								ProviderType: "enum",
								ReadOnly:     goutils.Pointer(true),
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

func TestUpsertMetadataCRM(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	errBadRequest := testutils.DataFromFile(t, "metadata/write/object15/bad-request.xml")
	payloadManyFields := testutils.DataFromFile(t, "metadata/write/object15/mix-of-many-fields-payload.xml")
	responseManyFields := testutils.DataFromFile(t, "metadata/write/object15/mix-of-many-fields-response.xml")
	payloadFieldPermissions := testutils.DataFromFile(t, "metadata/write/read-field-permissions-payload.xml")
	responseFieldPermissions := testutils.DataFromFile(t, "metadata/write/read-field-permissions-response.xml")
	payloadFieldPermissionsUpsert := testutils.DataFromFile(t, "metadata/write/write-field-permissions-payload.xml")
	responseFieldPermissionsUpsert := testutils.DataFromFile(t, "metadata/write/write-field-permissions-response.xml")
	responsePermissionSet := testutils.DataFromFile(t, "metadata/write/permission-set.json")
	responseUserInfo := testutils.DataFromFile(t, "metadata/write/user-info.json")
	duplicatePermissionAssignment := testutils.DataFromFile(t, "metadata/write/err-duplicate-permission-assignment.json")

	tests := []testroutines.UpsertMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFieldsMetadata},
		},
		{
			Name: "Upsert with invalid payload",
			Input: &common.UpsertMetadataParams{
				Fields: map[string][]common.FieldDefinition{
					"TestObject15__c": {
						{
							FieldName:   "IsReady__c",
							DisplayName: "IsReady",
							Description: "Indicates the readiness for next steps.",
							ValueType:   common.ValueTypeBoolean,
							Required:    true,
							Unique:      false,
							Indexed:     false,
							StringOptions: &common.StringFieldOptions{
								DefaultValue: goutils.Pointer("false"),
							},
						},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentXML(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/services/Soap/m/60.0"),
				},
				Then: mockserver.Response(http.StatusOK, errBadRequest),
			}.Server(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Can not specify 'required' for a CustomField of type Checkbox"), // nolint:err113
			},
		},
		{
			Name: "Upsert many fields of various types",
			Input: &common.UpsertMetadataParams{
				Fields: map[string][]common.FieldDefinition{
					"TestObject15__c": {
						{
							FieldName:   "Birthday__c",
							DisplayName: "Birthday",
							Description: "Story describing birthday",
							ValueType:   common.ValueTypeString,
							Required:    false,
							Unique:      false,
							Indexed:     false,
							StringOptions: &common.StringFieldOptions{
								Length: goutils.Pointer(30),
							},
						}, {
							FieldName:   "Hobby__c",
							DisplayName: "Hobby",
							Description: "Your hobby description",
							ValueType:   common.ValueTypeString,
							Required:    false,
							Unique:      false,
							Indexed:     false,
							StringOptions: &common.StringFieldOptions{
								Length:          goutils.Pointer(444),
								NumDisplayLines: goutils.Pointer(39),
							},
						}, {
							FieldName:   "Age__c",
							DisplayName: "Age",
							Description: "How many years you lived.",
							ValueType:   common.ValueTypeInt,
							Required:    true,
							Unique:      false,
							Indexed:     false,
							NumericOptions: &common.NumericFieldOptions{
								DefaultValue: goutils.Pointer(18.0),
								Precision:    goutils.Pointer(3),
								Scale:        goutils.Pointer(2),
							},
						}, {
							FieldName:   "Interests__c",
							DisplayName: "Interests",
							Description: "Topics that are of interest.",
							ValueType:   common.ValueTypeMultiSelect,
							Required:    true,
							Unique:      false,
							Indexed:     false,
							StringOptions: &common.StringFieldOptions{
								Values:           []string{"art", "travel", "swimming"},
								ValuesRestricted: true,
								DefaultValue:     goutils.Pointer("art"),
							},
						}, {
							FieldName:   "IsReady__c",
							DisplayName: "IsReady",
							Description: "Indicates the readiness for next steps.",
							ValueType:   common.ValueTypeBoolean,
							Required:    false,
							Unique:      false,
							Indexed:     false,
							StringOptions: &common.StringFieldOptions{
								DefaultValue: goutils.Pointer("false"),
							},
						}, {
							FieldName:   "Connection__c",
							DisplayName: "Connection",
							Description: "Connection to other objects.",
							ValueType:   common.ValueTypeOther,
							Required:    false,
							Unique:      false,
							Indexed:     false,
							Association: &common.AssociationDefinition{
								AssociationType: "associatedAccount",
								TargetObject:    "Account",
								// TargetField: "Identifier",  makes an IndirectLookup field
								// (Salesforce account must have that in the first place)
								OnDelete:               "SetNull",
								ReverseLookupFieldName: "MyAccount",
							},
						},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentXML(),
				Cases: mockserver.Cases{{
					// Upsert fields.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/services/Soap/m/60.0"),
						mockcond.BodyBytes(payloadManyFields),
					},
					Then: mockserver.Response(http.StatusOK, responseManyFields),
				}, {
					// Fetch permission set which contains field permissions.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/services/Soap/m/60.0"),
						mockcond.BodyBytes(payloadFieldPermissions),
					},
					Then: mockserver.Response(http.StatusOK, responseFieldPermissions),
				}, {
					// Upsert permission set with combined fields.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/services/Soap/m/60.0"),
						mockcond.BodyBytes(payloadFieldPermissionsUpsert),
					},
					Then: mockserver.Response(http.StatusOK, responseFieldPermissionsUpsert),
				}, {
					// Fetch permission set identifier.
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/services/data/v60.0/query"),
						mockcond.QueryParam("q",
							`SELECT Id,Name FROM PermissionSet WHERE Name='IntegrationCustomFieldVisibility'`),
					},
					Then: mockserver.ResponseChainedFuncs(
						mockserver.ContentJSON(),
						mockserver.Response(http.StatusOK, responsePermissionSet),
					),
				}, {
					// Fetch user identifier.
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/services/oauth2/userinfo"),
					},
					Then: mockserver.ResponseChainedFuncs(
						mockserver.ContentJSON(),
						mockserver.Response(http.StatusOK, responseUserInfo),
					),
				}, {
					// Assign permission set to the user.
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/services/data/v60.0/sobjects/PermissionSetAssignment"),
						mockcond.Body(`{"AssigneeId":"005006007008","PermissionSetId":"0PSak00000M9uBBGAZ"}`),
					},
					Then: mockserver.ResponseChainedFuncs(
						mockserver.ContentJSON(),
						mockserver.Response(http.StatusBadRequest, duplicatePermissionAssignment),
					),
				}},
			}.Server(),
			Expected: &common.UpsertMetadataResult{
				Success: true,
				Fields: map[string]map[string]common.FieldUpsertResult{
					"TestObject15__c": {
						"Birthday__c": {
							FieldName: "Birthday__c",
							Action:    "create",
						},
						"Hobby__c": {
							FieldName: "Hobby__c",
							Action:    "create",
						},
						"Age__c": {
							FieldName: "Age__c",
							Action:    "update",
						},
						"Interests__c": {
							FieldName: "Interests__c",
							Action:    "create",
						},
						"IsReady__c": {
							FieldName: "IsReady__c",
							Action:    "update",
						},
						"Connection__c": {
							FieldName: "Connection__c",
							Action:    "update",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			ctx := common.WithAuthToken(t.Context(), "TEST_ACCESS_TOKEN")

			tt.RunWithContext(t, ctx, func() (connectors.UpsertMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestUpsertMetadataNoAccessTokenCRM(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.UpsertMetadata{
		{
			Name: "Access token must be injected into the context",
			Input: &common.UpsertMetadataParams{
				Fields: map[string][]common.FieldDefinition{
					"Account": {
						{
							FieldName: "Birthday__c",
							ValueType: common.ValueTypeString,
						},
					},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingAccessToken},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.UpsertMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
