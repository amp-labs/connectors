package chilipiper

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	workspace := testutils.DataFromFile(t, "workspace.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.txt")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"workspace", "meme"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/workspace"),
					Then: mockserver.Response(http.StatusOK, workspace),
				}, {
					If:   mockcond.PathSuffix("/meme"),
					Then: mockserver.Response(http.StatusNotFound, unsupportedResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workspace": {
						DisplayName: "Workspace",
						FieldsMap: map[string]string{
							"emoji":    "emoji",
							"id":       "id",
							"logo":     "logo",
							"metadata": "metadata",
							"name":     "name",
						},
					},
				},
				Errors: map[string]error{
					"meme": mockutils.ExpectedSubsetErrors{
						common.ErrObjectNotSupported,
					},
				},
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
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	connector.BaseURL = serverURL

	return connector, nil
}
