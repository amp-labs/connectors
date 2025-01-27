package asana

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},

		{
			Name:         "Unknown object requested",
			Input:        []string{"groot"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{staticschema.ErrObjectNotFound},
		},

		{
			Name:       "Successfully describe multiple object with metadata",
			Input:      []string{"projects", "tags", "users", "workspaces"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"projects": {
						DisplayName: "Projects",
						FieldsMap: map[string]string{
							"gid":           "gid",
							"resource_type": "resource_type",
							"name":          "name",
						},
					},
					"tags": {
						DisplayName: "Tags",
						FieldsMap: map[string]string{
							"gid":           "gid",
							"resource_type": "resource_type",
							"name":          "name",
						},
					},
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"gid":           "gid",
							"resource_type": "resource_type",
							"name":          "name",
						},
					}, "workspaces": {
						DisplayName: "Workspaces",
						FieldsMap: map[string]string{
							"gid":           "gid",
							"resource_type": "resource_type",
							"name":          "name",
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
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(serverURL)

	return connector, nil
}
