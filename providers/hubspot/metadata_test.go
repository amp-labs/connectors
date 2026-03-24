package hubspot

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	metadataContactsProperties := testutils.DataFromFile(t, "metadata-contacts-properties-sampled.json")
	metadataContactsPipelines := testutils.DataFromFile(t, "metadata-contacts-external-pipelines.json")
	metadataDealsProperties := testutils.DataFromFile(t, "metadata-deals-properties-sampled.json")
	metadataDealsPipelines := testutils.DataFromFile(t, "metadata-deals-external-pipelines.json")
	metadataErrSchemaScopes := testutils.DataFromFile(t, "metadata-err-schemas-scope.json")
	responseLists := testutils.DataFromFile(t, "read-lists-1-first-page.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe contacts",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/crm/v3/properties/contacts"),
					Then: mockserver.Response(http.StatusOK, metadataContactsProperties),
				}, {
					// Connector will make this API call, output is valid empty data.
					// This is done to shrink the scope of a test. See tests below that focus on pipelines.
					If:   mockcond.Path("/crm/v3/pipelines/contacts"),
					Then: mockserver.ResponseString(http.StatusOK, "{}"),
				}, {
					// Real-world scenario doesn't require any fields for contacts.
					// For our mock unit test we require: "mobilephone".
					If:   mockcond.Path("/crm-object-schemas/v3/schemas/contacts"),
					Then: mockserver.ResponseString(http.StatusOK, `{"requiredProperties": ["mobilephone"]}`),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "contacts",
						Fields: map[string]common.FieldMetadata{
							// String
							"address": {
								DisplayName:  "Street Address",
								ValueType:    "string",
								ProviderType: "string.text",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"mobilephone": {
								DisplayName:  "Mobile Phone Number",
								ValueType:    "string",
								ProviderType: "string.phonenumber",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true), // required as per mock response.
								Values:       nil,
							},

							// Boolean.
							"hs_contact_enrichment_opt_out": {
								DisplayName:  "Enrichment opt out",
								ValueType:    "boolean",
								ProviderType: "bool.booleancheckbox",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"autogen": {
								DisplayName:  "autogen",
								ValueType:    "boolean",
								ProviderType: "enumeration.booleancheckbox",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(true),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},

							// Float.
							"associatedcompanyid": {
								DisplayName:  "Primary Associated Company ID",
								ValueType:    "float",
								ProviderType: "number.number",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"hubspotscore": {
								DisplayName:  "HubSpot Score",
								ValueType:    "float",
								ProviderType: "number.calculation_score",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"hs_associated_target_accounts": {
								DisplayName:  "Associated Target Accounts",
								ValueType:    "float",
								ProviderType: "number.calculation_rollup",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},

							// Single/Multi Select.
							"hs_content_membership_status": {
								DisplayName:  "Status",
								ValueType:    "singleSelect",
								ProviderType: "enumeration.select",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values: []common.FieldValue{{
									Value:        "active",
									DisplayValue: "Active",
								}, {
									Value:        "inactive",
									DisplayValue: "Inactive",
								}},
							},
							"hs_predictivecontactscorebucket": {
								DisplayName:  "Lead Rating",
								ValueType:    "singleSelect",
								ProviderType: "enumeration.radio",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values: []common.FieldValue{{
									Value:        "bucket_1",
									DisplayValue: "1 Star",
								}, {
									Value:        "bucket_2",
									DisplayValue: "2 Stars",
								}, {
									Value:        "bucket_3",
									DisplayValue: "3 Stars",
								}, {
									Value:        "bucket_4",
									DisplayValue: "4 Stars",
								}},
							},
							"hs_all_assigned_business_unit_ids": {
								DisplayName:  "Business units",
								ValueType:    "multiSelect",
								ProviderType: "enumeration.checkbox",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},

							// Datetime.
							"hs_first_subscription_create_date": {
								DisplayName:  "First subscription create date",
								ValueType:    "datetime",
								ProviderType: "datetime.calculation_rollup",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
							"hs_date_entered_customer": {
								DisplayName:  "Date entered 'Customer (Lifecycle Stage Pipeline)'",
								ValueType:    "datetime",
								ProviderType: "datetime.calculation_read_time",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},

							// Others.
							"hs_notes_last_activity": {
								DisplayName:  "Last Activity",
								ValueType:    "other",
								ProviderType: "object_coordinates.text",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"address":                         "Street Address",
							"associatedcompanyid":             "Primary Associated Company ID",
							"hs_predictivecontactscorebucket": "Lead Rating",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Describe contacts with external pipeline enumeration options",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/crm/v3/properties/contacts"),
					Then: mockserver.Response(http.StatusOK, metadataContactsProperties),
				}, {
					If:   mockcond.Path("/crm/v3/pipelines/contacts"),
					Then: mockserver.Response(http.StatusOK, metadataContactsPipelines),
				}, {
					// Required fields cannot be fetched. This is not a critical error.
					// In this case each field will be set to null indicating this info cannot be known.
					If:   mockcond.Path("/crm-object-schemas/v3/schemas/contacts"),
					Then: mockserver.Response(http.StatusForbidden, metadataErrSchemaScopes),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "contacts",
						Fields: map[string]common.FieldMetadata{
							// Values from External sources.
							"hs_pipeline": {
								DisplayName:  "Pipeline",
								ValueType:    "singleSelect",
								ProviderType: "enumeration.select",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   nil,
								Values: []common.FieldValue{{
									Value:        "contacts-lifecycle-pipeline",
									DisplayValue: "Lifecycle Stage Pipeline",
								}},
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Describe deals with external pipeline enumeration options",
			Input: []string{"deals"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/crm/v3/properties/deals"),
					Then: mockserver.Response(http.StatusOK, metadataDealsProperties),
				}, {
					If:   mockcond.Path("/crm/v3/pipelines/deals"),
					Then: mockserver.Response(http.StatusOK, metadataDealsPipelines),
				}, {
					If:   mockcond.Path("/crm-object-schemas/v3/schemas/deals"),
					Then: mockserver.ResponseString(http.StatusOK, `{}`),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"deals": {
						DisplayName: "deals",
						Fields: map[string]common.FieldMetadata{
							// Values from External sources.
							"pipeline": {
								DisplayName:  "Pipeline",
								ValueType:    "singleSelect",
								ProviderType: "enumeration.select",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values: []common.FieldValue{{
									Value:        "default",
									DisplayValue: "Sales Pipeline",
								}},
							},
							"dealstage": {
								DisplayName:  "Deal Stage",
								ValueType:    "singleSelect",
								ProviderType: "enumeration.radio",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								Values: []common.FieldValue{{
									Value:        "default:appointmentscheduled",
									DisplayValue: "Appointment Scheduled",
								}, {
									Value:        "default:qualifiedtobuy",
									DisplayValue: "Qualified To Buy",
								}, {
									Value:        "default:presentationscheduled",
									DisplayValue: "Presentation Scheduled",
								}, {
									Value:        "default:decisionmakerboughtin",
									DisplayValue: "Decision Maker Bought-In",
								}, {
									Value:        "default:contractsent",
									DisplayValue: "Contract Sent",
								}, {
									Value:        "default:closedwon",
									DisplayValue: "Closed Won",
								}, {
									Value:        "default:closedlost",
									DisplayValue: "Closed Lost",
								}},
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe lists, which is outside Properties API",
			Input: []string{"lists"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm/v3/lists/search"),
				Then:  mockserver.Response(http.StatusOK, responseLists),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lists": {
						DisplayName: "lists",
						Fields: map[string]common.FieldMetadata{
							"additionalProperties": {
								DisplayName:  "additionalProperties",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"updatedAt": {
								DisplayName:  "updatedAt",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
					},
				},
				Errors: nil,
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithModule(providers.ModuleHubspotCRM),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.providerInfo.BaseURL = mockutils.ReplaceURLOrigin(connector.providerInfo.BaseURL, serverURL)
	connector.moduleInfo.BaseURL = mockutils.ReplaceURLOrigin(connector.moduleInfo.BaseURL, serverURL)
	connector.crmAdapter.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.moduleInfo.BaseURL, serverURL))

	return connector, nil
}

func TestPersonaFieldDisplayValuePrefersDescription(t *testing.T) {
	t.Parallel()

	fdescription := fieldDescription{
		Name:      "HS_PERSONA",
		Type:      "enumeration",
		FieldType: "select",
		Options: []fieldEnumerationOption{
			{
				Value:       "persona-1",
				Label:       "Persona Label",
				Description: "Persona Description",
			},
			{
				Value: "persona-2",
				Label: "Persona Fallback",
			},
		},
	}

	metadata := fdescription.transformToFieldMetadata()

	if metadata.ValueType != common.ValueTypeSingleSelect {
		t.Fatalf("expected ValueType %q, got %q", common.ValueTypeSingleSelect, metadata.ValueType)
	}

	if len(metadata.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(metadata.Values))
	}

	if metadata.Values[0].DisplayValue != "Persona Description" {
		t.Fatalf("expected first display value to use description, got %q", metadata.Values[0].DisplayValue)
	}

	if metadata.Values[1].DisplayValue != "Persona Fallback" {
		t.Fatalf("expected second display value to fall back to label, got %q", metadata.Values[1].DisplayValue)
	}
}

func TestUpsertMetadataCRM(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	// Test scenario #1.
	payloadCreatedGroupName := testutils.DataFromFile(t, "custom/create/1-payload-create-property-group.json")
	responseCreatedGroupName := testutils.DataFromFile(t, "custom/create/2-response-create-property-group.json")
	payloadBatchCreateProperties1 := testutils.DataFromFile(t, "custom/create/3-payload-batch-create-properties.json")
	responseBatchCreateProperties1 := testutils.DataFromFile(t, "custom/create/4-response-batch-create-properties.json")

	// Test scenario #2.
	responseReadPropertyGroup := testutils.DataFromFile(t, "custom/update/1-read-property-group.json")
	payloadBatchCreateProperties2 := testutils.DataFromFile(t, "custom/update/2-payload-batch-create-properties.json")
	responseBatchCreateProperties2 := testutils.DataFromFile(t, "custom/update/3-response-batch-create-properties.json")
	payloadUpdateAge := testutils.DataFromFile(t, "custom/update/4-payload-update-property-age.json")
	responseUpdateAge := testutils.DataFromFile(t, "custom/update/5-response-update-property-age.json")
	payloadUpdateInterests := testutils.DataFromFile(t, "custom/update/6-payload-update-property-interests.json")
	responseUpdateInterests := testutils.DataFromFile(t, "custom/update/7-response-update-property-interests.json")

	tests := []testroutines.UpsertMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFieldsMetadata},
		},
		{
			Name: "Create properties in fresh new system",
			// Description:
			//		Group name does not exist and will be created.
			//		Then batch create "age" and "interest" fields.
			Input: &common.UpsertMetadataParams{
				Fields: map[string][]common.FieldDefinition{
					"Contact": {
						{
							FieldName:   "age__c",
							DisplayName: "Age",
							Description: "How many years you lived.",
							ValueType:   common.ValueTypeInt,
							Unique:      false,
						},
						{
							FieldName:   "interests__c",
							DisplayName: "Interests",
							Description: "Topics that are of interest.",
							ValueType:   common.ValueTypeMultiSelect,
							Unique:      false,
							StringOptions: &common.StringFieldOptions{
								Values: []string{"art", "travel", "swimming"},
							},
						},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{ // Group name does not exist.
						mockcond.MethodGET(),
						mockcond.Path("/crm/v3/properties/Contact/groups/integrationcreatedproperties"),
					},
					Then: mockserver.Response(http.StatusNotFound), // empty body.
				}, {
					If: mockcond.And{ // Create group name.
						mockcond.MethodPOST(),
						mockcond.Path("/crm/v3/properties/Contact/groups"),
						mockcond.BodyBytes(payloadCreatedGroupName),
					},
					Then: mockserver.Response(http.StatusCreated, responseCreatedGroupName),
				}, {
					If: mockcond.And{ // Create properties
						mockcond.MethodPOST(),
						mockcond.Path("/crm/v3/properties/Contact/batch/create"),
						mockcond.BodyBytes(payloadBatchCreateProperties1),
					},
					Then: mockserver.Response(http.StatusCreated, responseBatchCreateProperties1),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetUpsertMetadata,
			Expected: &common.UpsertMetadataResult{
				Success: true,
				Fields: map[string]map[string]common.FieldUpsertResult{
					"Contact": {
						"age__c": {
							FieldName: "age__c",
							Action:    "create",
							Metadata: map[string]any{
								"name":            "age__c",
								"label":           "Age",
								"type":            "number",
								"fieldType":       "number",
								"description":     "How many years you lived.",
								"groupName":       "integrationcreatedproperties",
								"options":         []any{},
								"displayOrder":    float64(-1),
								"calculated":      false,
								"externalOptions": false,
								"hasUniqueValue":  false,
								"hidden":          false,
								"formField":       true,
								"dataSensitivity": "non_sensitive",
							},
						},
						"interests__c": {
							FieldName: "interests__c",
							Action:    "create",
							Metadata: map[string]any{
								"name":        "interests__c",
								"label":       "Interests",
								"type":        "enumeration",
								"fieldType":   "select",
								"description": "Topics that are of interest.",
								"groupName":   "integrationcreatedproperties",
								"options": []any{
									map[string]any{
										"label":        "art",
										"value":        "art",
										"description":  "art",
										"displayOrder": float64(3),
										"hidden":       false,
									},
									map[string]any{
										"label":        "travel",
										"value":        "travel",
										"description":  "travel",
										"displayOrder": float64(1),
										"hidden":       false,
									},
									map[string]any{
										"label":        "swimming",
										"value":        "swimming",
										"description":  "swimming",
										"displayOrder": float64(2),
										"hidden":       false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "Update existing properties and create some new",
			// Description:
			//		Group name already exists and will be retrieved.
			//		Batch create will fail for "hobby" and "is-ready" but will be ok for "age" and "interests".
			//		Dedicated calls will be made to update "age" and to update "interests".
			Input: &common.UpsertMetadataParams{
				Fields: map[string][]common.FieldDefinition{
					"Contact": {
						{
							FieldName:   "hobby__c",
							DisplayName: "Hobby",
							Description: "Your hobby description",
							ValueType:   common.ValueTypeString,
							Unique:      true,
						},
						{
							FieldName:   "age__c",
							DisplayName: "Age",
							Description: "How old are you?",
							ValueType:   common.ValueTypeInt,
							Unique:      false,
						},
						{
							FieldName:   "interests__c",
							DisplayName: "Interests",
							Description: "Topics that are of interest.",
							ValueType:   common.ValueTypeMultiSelect,
							Unique:      false,
							StringOptions: &common.StringFieldOptions{
								Values: []string{"art", "travel", "swimming"},
							},
						},
						{
							FieldName:   "isready__c",
							DisplayName: "IsReady",
							Description: "Indicates the readiness for next steps.",
							ValueType:   common.ValueTypeBoolean,
							Unique:      false,
						},
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{ // Group name is fetched.
						mockcond.MethodGET(),
						mockcond.Path("/crm/v3/properties/Contact/groups/integrationcreatedproperties"),
					},
					Then: mockserver.Response(http.StatusOK, responseReadPropertyGroup),
				}, {
					If: mockcond.And{ // Create properties
						mockcond.MethodPOST(),
						mockcond.Path("/crm/v3/properties/Contact/batch/create"),
						mockcond.BodyBytes(payloadBatchCreateProperties2),
					},
					Then: mockserver.Response(http.StatusMultiStatus, responseBatchCreateProperties2),
				}, {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/crm/v3/properties/Contact/age__c"),
						mockcond.BodyBytes(payloadUpdateAge),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdateAge),
				}, {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/crm/v3/properties/Contact/interests__c"),
						mockcond.BodyBytes(payloadUpdateInterests),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdateInterests),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetUpsertMetadata,
			Expected: &common.UpsertMetadataResult{
				Success: true,
				Fields: map[string]map[string]common.FieldUpsertResult{
					"Contact": {
						"age__c": {
							FieldName: "age__c",
							Action:    "update",
							Metadata: map[string]any{
								"label":       "Age",
								"description": "How old are you?",
							},
						},
						"interests__c": {
							FieldName: "interests__c",
							Action:    "update",
							Metadata: map[string]any{
								"label":       "Interests",
								"description": "Topics that are of interest.",
							},
						},
						"isready__c": {
							FieldName: "isready__c",
							Action:    "create",
							Metadata: map[string]any{
								"name":        "isready__c",
								"label":       "IsReady",
								"type":        "bool",
								"fieldType":   "booleancheckbox",
								"description": "Indicates the readiness for next steps.",
								"groupName":   "integrationcreatedproperties",
							},
						},
						"hobby__c": {
							FieldName: "hobby__c",
							Action:    "create",
							Metadata: map[string]any{
								"name":        "hobby__c",
								"label":       "Hobby",
								"type":        "string",
								"fieldType":   "text",
								"description": "Your hobby description",
								"groupName":   "integrationcreatedproperties",
							},
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
