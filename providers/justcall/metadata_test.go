package justcall

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for Users and Calls",
			Input:      []string{"users", "calls"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "ID",
								ValueType:    "integer",
								ProviderType: "integer",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"calls": {
						DisplayName: "Calls",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "ID",
								ValueType:    "integer",
								ProviderType: "integer",
							},
							"direction": {
								DisplayName:  "Direction",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "Status",
								ValueType:    "string",
								ProviderType: "string",
							},
							"duration": {
								DisplayName:  "Duration",
								ValueType:    "integer",
								ProviderType: "integer",
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
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
