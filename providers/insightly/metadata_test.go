package insightly

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

	responseContacts := testutils.DataFromFile(t, "read/contacts/custom-fields.json")
	responseFruitObject := testutils.DataFromFile(t, "read/custom-objects/fruit.json")
	responseFruitFields := testutils.DataFromFile(t, "read/fruits-custom-object/custom-fields.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:  "Successfully return metadata for Teams",
			Input: []string{"Teams"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3.1/CustomFields/Teams"),
				Then:  mockserver.ResponseString(http.StatusOK, `[]`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Teams": {
						DisplayName: "Teams",
						Fields: map[string]common.FieldMetadata{
							"TEAMMEMBERS": {
								DisplayName:  "TEAMMEMBERS",
								ValueType:    "other",
								ProviderType: "array",
							},
							"TEAM_ID": {
								DisplayName:  "TEAM_ID",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"TEAM_NAME": {
								DisplayName:  "TEAM_NAME",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: nil,
			},
		},
		{
			Name:  "Successfully return metadata for Contacts with custom fields",
			Input: []string{"Contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3.1/CustomFields/Contacts"),
				Then:  mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							// Metadata from schema.json."butterflies": mockutils.ExpectedSubsetErrors{
							"FIRST_NAME": {
								DisplayName:  "FIRST_NAME",
								ValueType:    "string",
								ProviderType: "string",
							},
							// Custom fields.
							"Hobby__c": {
								DisplayName:  "Hobby",
								ValueType:    "string",
								ProviderType: "TEXT",
								ReadOnly:     goutils.Pointer(false),
							},
							"Interests__c": {
								DisplayName:  "Interests",
								ValueType:    "multiSelect",
								ProviderType: "MULTISELECT",
								ReadOnly:     goutils.Pointer(false),
								Values: common.FieldValues{{
									Value:        "3",
									DisplayValue: "Art",
								}, {
									Value:        "6",
									DisplayValue: "Food",
								}, {
									Value:        "5",
									DisplayValue: "Music",
								}, {
									Value:        "2",
									DisplayValue: "Sports",
								}, {
									Value:        "1",
									DisplayValue: "Technology",
								}, {
									Value:        "4",
									DisplayValue: "Travel",
								}},
							},
							"Newsletter_Subscription__c": {
								DisplayName:  "Newsletter Subscription",
								ValueType:    "other",
								ProviderType: "BIT",
								ReadOnly:     goutils.Pointer(false),
							},
							"Preferred_Contact_Method__c": {
								DisplayName:  "Preferred Contact Method",
								ValueType:    "singleSelect",
								ProviderType: "DROPDOWN",
								ReadOnly:     goutils.Pointer(false),
								Values: common.FieldValues{{
									Value:        "1",
									DisplayValue: "Email",
								}, {
									Value:        "2",
									DisplayValue: "Phone",
								}, {
									Value:        "3",
									DisplayValue: "SMS",
								}, {
									Value:        "4",
									DisplayValue: "WhatsApp",
								}},
							},
							"Date1__c": {
								DisplayName:  "Date1",
								ValueType:    "date",
								ProviderType: "DATE",
								ReadOnly:     goutils.Pointer(false),
							},
							"Date2__c": {
								DisplayName:  "Date2",
								ValueType:    "datetime",
								ProviderType: "DATETIME",
								ReadOnly:     goutils.Pointer(false),
							},
							"MultiText1__c": {
								DisplayName:  "MultiText1",
								ValueType:    "string",
								ProviderType: "MULTILINETEXT",
								ReadOnly:     goutils.Pointer(false),
							},
							"Number1__c": {
								DisplayName:  "Number1",
								ValueType:    "float",
								ProviderType: "NUMERIC",
								ReadOnly:     goutils.Pointer(false),
							},
							"Percent1__c": {
								DisplayName:  "Percent1",
								ValueType:    "float",
								ProviderType: "PERCENT",
								ReadOnly:     goutils.Pointer(false),
							},
							"AutoNumber1__c": {
								DisplayName:  "AutoNumber1",
								ValueType:    "other",
								ProviderType: "AUTONUMBER",
								ReadOnly:     goutils.Pointer(true),
							},
						},
					},
				},
				Errors: nil,
			},
		},
		{
			Name:  "Successfully return metadata for custom object Fruits with custom fields",
			Input: []string{"Fruit__c"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3.1/CustomObjects/Fruit__c"),
					Then: mockserver.Response(http.StatusOK, responseFruitObject),
				}, {
					If:   mockcond.Path("/v3.1/CustomFields/Fruit__c"),
					Then: mockserver.Response(http.StatusOK, responseFruitFields),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Fruit__c": {
						DisplayName: "Fruits",
						Fields: map[string]common.FieldMetadata{
							// Properties which are always common across all custom object types.
							"RECORD_NAME": {
								DisplayName:  "RECORD_NAME",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"CUSTOMFIELDS": {
								DisplayName:  "CUSTOMFIELDS",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
							},
							// Custom fields.
							"Weight__c": {
								DisplayName:  "Weight",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "NUMERIC",
								ReadOnly:     goutils.Pointer(false),
							},
							"Color__c": {
								DisplayName:  "Color",
								ValueType:    common.ValueTypeString,
								ProviderType: "TEXT",
								ReadOnly:     goutils.Pointer(false),
							},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
