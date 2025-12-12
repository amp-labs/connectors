package cloudtalk

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for Calls",
			Input:      []string{"calls"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"calls": {
						DisplayName: "Calls",
						Fields: map[string]common.FieldMetadata{
							"Agent": {
								DisplayName:  "Agent",
								ValueType:    "other",
								ProviderType: "object",
							},
							"Contact": {
								DisplayName:  "Contact",
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
			Module:              common.ModuleRoot,
			AuthenticatedClient: &http.Client{},
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	// Override the base URL to point to the test server
	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
