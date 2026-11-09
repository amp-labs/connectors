package wordpress

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:       "Successful metadata for multiple objects",
			Input:      []string{"posts", "pages"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"posts": {
						DisplayName: "Posts",
						Fields: map[string]common.FieldMetadata{
							"date": {
								DisplayName:  "Date",
								ProviderType: "null|string(date-time)",
								ValueType:    "datetime",
							},
							"comment_status": {
								DisplayName:  "Comment Status",
								ProviderType: "string",
								ValueType:    "string",
								Values: common.FieldValues{{
									Value:        "open",
									DisplayValue: "open",
								}, {
									Value:        "closed",
									DisplayValue: "closed",
								}},
							},
						},
					},
					"pages": {
						DisplayName: "Pages",
						Fields: map[string]common.FieldMetadata{
							"jetpack_sharing_enabled": {
								DisplayName:  "Jetpack Sharing Enabled",
								ProviderType: "boolean",
								ValueType:    "boolean",
							},
							"link": {
								DisplayName:  "Link",
								ProviderType: "string(uri)",
								ValueType:    "string",
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
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
		Workspace:           "test-workspace",
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(server.URL)

	return connector, nil
}
