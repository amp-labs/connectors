package chilipiper

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	workspace := testutils.DataFromFile(t, "workspace.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.txt")
	emptyResponse := testutils.DataFromFile(t, "empty.json")

	tests := []testconn.TestCaseListObjectMetadata{
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
					If:   mockcond.Path("/api/fire-edge/v1/org/workspace"),
					Then: mockserver.Response(http.StatusOK, workspace),
				}, {
					If:   mockcond.Path("/api/fire-edge/v1/org/meme"),
					Then: mockserver.Response(http.StatusNotFound, unsupportedResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workspace": {
						DisplayName: "Workspace",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"name": {
								DisplayName: "name",
								ValueType:   common.ValueTypeString,
							},
							"logo": {
								DisplayName: "logo",
								ValueType:   common.ValueTypeOther,
							},
							"metadata": {
								DisplayName: "metadata",
								ValueType:   common.ValueTypeOther,
							},
						},
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
					"meme": common.ErrObjectNotSupported,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Empty response falls back to static schema",
			Input: []string{"workspace/users"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/fire-edge/v1/org/workspace/users"),
					Then: mockserver.Response(http.StatusOK, emptyResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workspace/users": {
						DisplayName: "workspace/users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"isSuperAdmin": {
								DisplayName:  "isSuperAdmin",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
							"licenses": {
								DisplayName:  "licenses",
								ValueType:    common.ValueTypeOther,
								ProviderType: "object",
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"isSuperAdmin": "isSuperAdmin",
							"licenses":     "licenses",
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

	connector.BaseURL = serverURL
	connector.BaseURL = mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL)

	return connector, nil
}
