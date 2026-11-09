package zoom

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
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
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"godzilla": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name: "Successfully describe multiple objects with metadata",
			Input: []string{
				"users", "groups",
				"activities_report", "devices_groups",
				"meetings", "account_meetings"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"groups": {
						DisplayName: "Groups",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
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
					},
					"meetings": {
						DisplayName: "My Meetings",
						Fields: map[string]common.FieldMetadata{
							"duration": {
								DisplayName:  "duration",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
					"account_meetings": {
						DisplayName: "All Meetings",
						Fields: map[string]common.FieldMetadata{
							"duration": {
								DisplayName:  "duration",
								ValueType:    "int",
								ProviderType: "integer",
							},
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
			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(server.URL)

	return connector, nil
}
