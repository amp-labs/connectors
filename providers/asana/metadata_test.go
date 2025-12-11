package asana

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseProjects := testutils.DataFromFile(t, "read-projects.json")
	responseSingleProject := testutils.DataFromFile(t, "single-project.json")
	responseUsers := testutils.DataFromFile(t, "read-users.json")
	responseSingleUser := testutils.DataFromFile(t, "single-user.json")
	responseTags := testutils.DataFromFile(t, "read-tags.json")
	responseSingleTag := testutils.DataFromFile(t, "single-tag.json")

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},

		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"projects", "tags", "users", "workspaces"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/api/1.0/projects/12345"),
						Then: mockserver.Response(200, responseSingleProject),
					},
					{
						If:   mockcond.Path("/api/1.0/projects"),
						Then: mockserver.Response(200, responseProjects),
					},
					{
						If:   mockcond.Path("/api/1.0/tags"),
						Then: mockserver.Response(200, responseTags),
					},
					{
						If:   mockcond.Path("/api/1.0/tags/12225"),
						Then: mockserver.Response(200, responseSingleTag),
					},
					{
						If:   mockcond.Path("/api/1.0/users"),
						Then: mockserver.Response(200, responseUsers),
					},
					{
						If:   mockcond.Path("/api/1.0/users/1245"),
						Then: mockserver.Response(200, responseSingleUser),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"projects": {
						DisplayName: "Projects",
						Fields: map[string]common.FieldMetadata{
							"gid": {
								DisplayName:  "gid",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"resource_type": {
								DisplayName:  "resource_type",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
						},
					},
					"tags": {
						DisplayName: "Tags",
						Fields: map[string]common.FieldMetadata{
							"gid": {
								DisplayName:  "gid",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"resource_type": {
								DisplayName:  "resource_type",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"gid": {
								DisplayName:  "gid",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"resource_type": {
								DisplayName:  "resource_type",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
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
