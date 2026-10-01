package confluence

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Successfully describe one object with metadata",
			Input:      []string{"pages", "blogposts"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"pages": {
						DisplayName: "Pages",
						Fields: map[string]common.FieldMetadata{
							"parentType": {
								DisplayName:  "parentType",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{{
									Value:        "page",
									DisplayValue: "page",
								}, {
									Value:        "whiteboard",
									DisplayValue: "whiteboard",
								}, {
									Value:        "database",
									DisplayValue: "database",
								}, {
									Value:        "embed",
									DisplayValue: "embed",
								}, {
									Value:        "folder",
									DisplayValue: "folder",
								}},
							},
						},
					},
					"blogposts": {
						DisplayName: "Blog Posts",
						Fields: map[string]common.FieldMetadata{
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
							},
							"version": {
								DisplayName:  "version",
								ValueType:    "other",
								ProviderType: "object",
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
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestAdapter(tt.Server)
			})
		})
	}
}

func constructTestAdapter(server *httptest.Server) (*Adapter, error) {
	base, err := components.NewBaseConnector(providers.Atlassian, common.ConnectorParams{
		Module:              providers.ModuleAtlassianConfluence,
		AuthenticatedClient: server.Client(),
		Workspace:           "test-workspace",
		Metadata: map[string]string{
			"cloudId": "ebc887b2-7e61-4059-ab35-71f15cc16e12", // any random value will work for the test
		},
	})
	if err != nil {
		return nil, err
	}

	base.SetUnitTestMockServerBaseUrl(server.URL)

	return NewAdapter(base)
}
