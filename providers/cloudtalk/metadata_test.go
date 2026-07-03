package cloudtalk

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:       "Successful metadata for Calls",
			Input:      []string{"calls"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
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

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: server.Client(),
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	// Override the base URL to point to the test server
	connector.SetUnitTestBaseURL(server.URL)

	return connector, nil
}
