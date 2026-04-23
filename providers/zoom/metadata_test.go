package zoom

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"godzilla"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"godzilla": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"users", "groups"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"display_name": "display_name",
							"dept":         "dept",
							"email":        "email",
							"status":       "status",
						},
					},
					"groups": {
						DisplayName: "Groups",
						FieldsMap: map[string]string{
							"name": "name",
							"id":   "id",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"activities_report", "devices_groups"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities_report": {
						DisplayName: "Activities Report",
						Fields: map[string]common.FieldMetadata{
							"type": {
								DisplayName:  "type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{{
									Value:        "Sign in",
									DisplayValue: "Sign in",
								}, {
									Value:        "Sign out",
									DisplayValue: "Sign out",
								}},
							},
						},
						FieldsMap: map[string]string{
							"client_type": "client_type",
							"type":        "type",
							"email":       "email",
							"version":     "version",
						},
					},
					"devices_groups": {
						DisplayName: "Devices Groups",
						Fields: map[string]common.FieldMetadata{
							"zdm_group_id": {
								DisplayName:  "zdm_group_id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
						FieldsMap: map[string]string{
							"name":        "name",
							"description": "description",
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
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
