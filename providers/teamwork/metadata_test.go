package teamwork

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
			Name:       "Successful metadata for projects and tasks",
			Input:      []string{"projects", "tasks"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"projects": {
						DisplayName: "Projects",
						Fields: map[string]common.FieldMetadata{
							"company": {
								DisplayName:  "Company",
								ValueType:    "other",
								ProviderType: "object",
							},
							"tags": {
								DisplayName:  "Tags",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
					"tasks": {
						DisplayName: "Tasks",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"description": {
								DisplayName:  "Description",
								ValueType:    "string",
								ProviderType: "string",
							},
							"card": {
								DisplayName:  "Card",
								ValueType:    "other",
								ProviderType: "object",
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
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
