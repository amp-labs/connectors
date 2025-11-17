package copper

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCustomFields := testutils.DataFromFile(t, "custom/fields.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successful metadata for Projects and Leads",
			Input: []string{"projects", "leads", "companies"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/developer_api/v1/custom_field_definitions"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"projects": {
						DisplayName: "Projects",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"leads": {
						DisplayName: "Leads",
						Fields: map[string]common.FieldMetadata{
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
							},
							"custom_field_fruits": {
								DisplayName:  "Fruits",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "Dropdown",
								Values: []common.FieldValue{{
									Value:        "2082340",
									DisplayValue: "Banana",
								}, {
									Value:        "2082341",
									DisplayValue: "Strawberries",
								}},
							},
						},
					},
					"companies": {
						DisplayName: "Companies",
						Fields: map[string]common.FieldMetadata{
							"custom_field_birthday": {
								DisplayName:  "Birthday",
								ValueType:    "datetime",
								ProviderType: "Date",
							},
							"custom_field_child_of": {
								DisplayName:  "Child of",
								ValueType:    "other",
								ProviderType: "Connect",
							},
							"custom_field_favnum": {
								DisplayName:  "FavNum",
								ValueType:    "float",
								ProviderType: "Float",
							},
							"custom_field_fruits": {
								DisplayName:  "Fruits",
								ValueType:    "singleSelect",
								ProviderType: "Dropdown",
								Values: []common.FieldValue{{
									Value:        "2082340",
									DisplayValue: "Banana",
								}, {
									Value:        "2082341",
									DisplayValue: "Strawberries",
								}},
							},
							"custom_field_hryvnia": {
								DisplayName:  "Hryvnia",
								ValueType:    "float",
								ProviderType: "Currency",
							},
							"custom_field_isbouillonsoup": {
								DisplayName:  "IsBouillonSoup",
								ValueType:    "boolean",
								ProviderType: "Checkbox",
							},
							"custom_field_many": {
								DisplayName:  "Many",
								ValueType:    "multiSelect",
								ProviderType: "MultiSelect",
								Values: []common.FieldValue{{
									Value:        "2082480",
									DisplayValue: "Option 1",
								}, {
									Value:        "2082481",
									DisplayValue: "Option 2",
								}},
							},
							"custom_field_mywebsite": {
								DisplayName:  "MyWebsite",
								ValueType:    "string",
								ProviderType: "URL",
							},
							"custom_field_parent_of": {
								DisplayName:  "Parent of",
								ValueType:    "other",
								ProviderType: "Connect",
							},
							"custom_field_progression": {
								DisplayName:  "Progression",
								ValueType:    "float",
								ProviderType: "Percentage",
							},
							"custom_field_story": {
								DisplayName:  "Story",
								ValueType:    "string",
								ProviderType: "Text",
							},
							"custom_field_textentry": {
								DisplayName:  "TextEntry",
								ValueType:    "string",
								ProviderType: "String",
							},
							"custom_fields": {
								DisplayName:  "custom_fields",
								ValueType:    "other",
								ProviderType: "other",
							},
							"date_created": {
								DisplayName:  "date_created",
								ValueType:    "float",
								ProviderType: "float",
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
			Metadata: map[string]string{
				"userEmail": "john@test.com",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
