package bambooHR

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name: "Successful metadata for core HR objects",
			Input: []string{
				"employees_directory",
				"meta_users",
				"employees",
				"time_off_requests",
				"whos_out",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"employees_directory": {
						DisplayName: "Employees Directory",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"firstName": {
								DisplayName:  "First Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"meta_users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"employees": {
						DisplayName: "Employees",
						Fields: map[string]common.FieldMetadata{
							"employeeId": {
								DisplayName:  "employeeId",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"time_off_requests": {
						DisplayName: "Time Off Requests",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
					"whos_out": {
						DisplayName: "Who's Out",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
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
			Name:         "No objects returns missing objects error",
			Input:        nil,
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"employees", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"employees": {
						DisplayName: "Employees",
						Fields: map[string]common.FieldMetadata{
							"employeeId": {
								DisplayName:  "employeeId",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"unknown_object": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(serverURL)

	return connector, nil
}
