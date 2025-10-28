package sellsy

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for Tasks and Favourite Filters",
			Input:      []string{"tasks", "companies/favourite-filters"},
			Server:     mockserver.Dummy(),
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
