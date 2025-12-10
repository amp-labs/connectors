package justcall

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
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
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"role": {
								DisplayName:  "role",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"calls": {
						DisplayName: "Calls",
						Fields: map[string]common.FieldMetadata{
							"contact_number": {
								DisplayName:  "contact_number",
								ValueType:    "string",
								ProviderType: "string",
							},
							"agent_name": {
								DisplayName:  "agent_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"call_date": {
								DisplayName:  "call_date",
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
	connector.BaseURL = serverURL

	return connector, nil
}
