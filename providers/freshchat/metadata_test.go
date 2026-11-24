package freshchat

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object returns ErrObjectNotSupported",
			Input:      []string{"nonexistent"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent": common.ErrObjectNotSupported,
				},
			},
		},
		{
			// Metadata is served statically from the embedded schema,
			// so no provider call is made (Dummy server).
			Name:       "Describe users and agents with their field types",
			Input:      []string{"users", "agents"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users List",
						Fields: common.FieldsMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"created_time": {
								DisplayName:  "created_time",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "date-time",
							},
						},
					},
					"agents": {
						DisplayName: "Agents List",
						Fields: common.FieldsMetadata{
							"is_deactivated": {
								DisplayName:  "is_deactivated",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
							"agent_capacity": {
								DisplayName:  "agent_capacity",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return NewConnector(common.ConnectorParams{
					AuthenticatedClient: mockutils.NewClient(),
					Workspace:           "test",
				})
			})
		})
	}
}
