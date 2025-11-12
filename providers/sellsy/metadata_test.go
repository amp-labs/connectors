package sellsy

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

	responseCustomFields := testutils.DataFromFile(t, "read/custom-fields.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successful metadata for Tasks and Favourite Filters",
			Input: []string{"tasks", "companies/favourite-filters"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tasks": {
						DisplayName: "Tasks",
						Fields: map[string]common.FieldMetadata{
							"priority": {
								DisplayName:  "priority",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{{
									Value:        "todo",
									DisplayValue: "todo",
								}, {
									Value:        "done",
									DisplayValue: "done",
								}},
							},
						},
					},
					"companies/favourite-filters": {
						DisplayName: "Companies Favourite Filters",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"type": {
								DisplayName:  "type",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Custom fields are returned as part of Contacts metadata",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/custom-fields"),
				Then:  mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"age": {
								DisplayName:  "Age",
								ValueType:    "int",
								ProviderType: "numeric",
								Values:       nil,
							},
							"fruits": {
								DisplayName:  "Fruits",
								ValueType:    "singleSelect",
								ProviderType: "radio",
								Values: common.FieldValues{{
									Value:        "9",
									DisplayValue: "Orange",
								}, {
									Value:        "10",
									DisplayValue: "Strawberry",
								}, {
									Value:        "11",
									DisplayValue: "Kiwi",
								}},
							},
							"hobbies": {
								DisplayName:  "Hobbies",
								ValueType:    "multiSelect",
								ProviderType: "checkbox",
								Values: common.FieldValues{{
									Value:        "12",
									DisplayValue: "Art",
								}, {
									Value:        "13",
									DisplayValue: "Travelling",
								}, {
									Value:        "14",
									DisplayValue: "Movies",
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
