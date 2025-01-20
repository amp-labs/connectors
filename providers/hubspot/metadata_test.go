package hubspot

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	metadataContactsProperties := testutils.DataFromFile(t, "metadata-contacts-properties-sampled.json")
	metadataContactsPipelines := testutils.DataFromFile(t, "metadata-contacts-external-pipelines.json")
	metadataDealsProperties := testutils.DataFromFile(t, "metadata-deals-properties-sampled.json")
	metadataDealsPipelines := testutils.DataFromFile(t, "metadata-deals-external-pipelines.json")
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
					If:   mockcond.PathSuffix("/crm/v3/properties/contacts"),
					Then: mockserver.Response(http.StatusOK, metadataContactsProperties),
				}, {
					// Connector will make this API call, output is valid empty data.
					// This is done to shrink the scope of a test. See tests below that focus on pipelines.
					If:   mockcond.PathSuffix("/crm/v3/pipelines/contacts"),
					Then: mockserver.ResponseString(http.StatusOK, "{}"),
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
								ReadOnly:     false,
								Values:       nil,
							},
							"mobilephone": {
								DisplayName:  "Mobile Phone Number",
								ValueType:    "string",
								ProviderType: "string.phonenumber",
								ReadOnly:     false,
								Values:       nil,
							},

							// Boolean.
							"hs_contact_enrichment_opt_out": {
								DisplayName:  "Enrichment opt out",
								ValueType:    "boolean",
								ProviderType: "bool.booleancheckbox",
								ReadOnly:     true,
								Values:       nil,
							},
							"autogen": {
								DisplayName:  "autogen",
								ValueType:    "boolean",
								ProviderType: "enumeration.booleancheckbox",
								ReadOnly:     false,
								Values:       nil,
							},

							// Float.
							"associatedcompanyid": {
								DisplayName:  "Primary Associated Company ID",
								ValueType:    "float",
								ProviderType: "number.number",
								ReadOnly:     false,
								Values:       nil,
							},
							"hubspotscore": {
								DisplayName:  "HubSpot Score",
								ValueType:    "float",
								ProviderType: "number.calculation_score",
								ReadOnly:     true,
								Values:       nil,
							},
							"hs_associated_target_accounts": {
								DisplayName:  "Associated Target Accounts",
								ValueType:    "float",
								ProviderType: "number.calculation_rollup",
								ReadOnly:     true,
								Values:       nil,
							},

							// Single/Multi Select.
							"hs_content_membership_status": {
								DisplayName:  "Status",
								ValueType:    "singleSelect",
								ProviderType: "enumeration.select",
								ReadOnly:     false,
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
								ReadOnly:     true,
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
								ReadOnly:     false,
								Values:       nil,
							},

							// Datetime.
							"hs_first_subscription_create_date": {
								DisplayName:  "First subscription create date",
								ValueType:    "datetime",
								ProviderType: "datetime.calculation_rollup",
								ReadOnly:     true,
								Values:       nil,
							},
							"hs_date_entered_customer": {
								DisplayName:  "Date entered 'Customer (Lifecycle Stage Pipeline)'",
								ValueType:    "datetime",
								ProviderType: "datetime.calculation_read_time",
								ReadOnly:     true,
								Values:       nil,
							},

							// Others.
							"hs_notes_last_activity": {
								DisplayName:  "Last Activity",
								ValueType:    "other",
								ProviderType: "object_coordinates.text",
								ReadOnly:     true,
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
					If:   mockcond.PathSuffix("/crm/v3/properties/contacts"),
					Then: mockserver.Response(http.StatusOK, metadataContactsProperties),
				}, {
					If:   mockcond.PathSuffix("/crm/v3/pipelines/contacts"),
					Then: mockserver.Response(http.StatusOK, metadataContactsPipelines),
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
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "subscriber",
									DisplayValue: "Subscriber",
								}, {
									Value:        "lead",
									DisplayValue: "Lead",
								}, {
									Value:        "marketingqualifiedlead",
									DisplayValue: "Marketing Qualified Lead",
								}, {
									Value:        "salesqualifiedlead",
									DisplayValue: "Sales Qualified Lead",
								}, {
									Value:        "opportunity",
									DisplayValue: "Opportunity",
								}, {
									Value:        "customer",
									DisplayValue: "Customer",
								}, {
									Value:        "evangelist",
									DisplayValue: "Evangelist",
								}, {
									Value:        "other",
									DisplayValue: "Other",
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
					If:   mockcond.PathSuffix("/crm/v3/properties/deals"),
					Then: mockserver.Response(http.StatusOK, metadataDealsProperties),
				}, {
					If:   mockcond.PathSuffix("/crm/v3/pipelines/deals"),
					Then: mockserver.Response(http.StatusOK, metadataDealsPipelines),
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
								ReadOnly:     false,
								Values: []common.FieldValue{{
									Value:        "appointmentscheduled",
									DisplayValue: "Appointment Scheduled",
								}, {
									Value:        "qualifiedtobuy",
									DisplayValue: "Qualified To Buy",
								}, {
									Value:        "presentationscheduled",
									DisplayValue: "Presentation Scheduled",
								}, {
									Value:        "decisionmakerboughtin",
									DisplayValue: "Decision Maker Bought-In",
								}, {
									Value:        "contractsent",
									DisplayValue: "Contract Sent",
								}, {
									Value:        "closedwon",
									DisplayValue: "Closed Won",
								}, {
									Value:        "closedlost",
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
			Name:  "Successfully describe lists, which is outside ObjectsAPI",
			Input: []string{"lists"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/v3/lists/search"),
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
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"updatedAt": {
								DisplayName:  "updatedAt",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
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
		WithAuthenticatedClient(http.DefaultClient),
		WithModule(ModuleCRM),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
