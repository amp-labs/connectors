package notion

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
			Name: "Successful metadata for core Notion objects",
			Input: []string{
				"users",
				"pages",
				"data_sources",
				"blocks",
				"file_uploads",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"type": {
								DisplayName:  "type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "bot", DisplayValue: "bot"},
									{Value: "person", DisplayValue: "person"},
								},
							},
						},
					},
					"pages": {
						DisplayName: "Pages",
						Fields: map[string]common.FieldMetadata{
							"url": {
								DisplayName:  "url",
								ValueType:    "string",
								ProviderType: "string",
							},
							"properties": {
								DisplayName:  "properties",
								ValueType:    "other",
								ProviderType: "object",
							},
						},
					},
					"data_sources": {
						DisplayName: "Data Sources",
						Fields: map[string]common.FieldMetadata{
							"title": {
								DisplayName:  "title",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
					"blocks": {
						DisplayName: "Blocks",
						Fields: map[string]common.FieldMetadata{
							"has_children": {
								DisplayName:  "has_children",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"file_uploads": {
						DisplayName: "File Uploads",
						Fields: map[string]common.FieldMetadata{
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "expired", DisplayValue: "expired"},
									{Value: "failed", DisplayValue: "failed"},
									{Value: "pending", DisplayValue: "pending"},
									{Value: "uploaded", DisplayValue: "uploaded"},
								},
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"users", "search"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"search": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
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
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
