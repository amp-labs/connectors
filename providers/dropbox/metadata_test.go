package dropbox

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
			Name: "Successful metadata for Dropbox list objects",
			Input: []string{
				"files",
				"file_requests",
				"shared_links",
				"shared_folders",
				"mountable_shared_folders",
				"received_files",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"files": {
						DisplayName: "Files",
						Fields: map[string]common.FieldMetadata{
							".tag": {
								DisplayName:  "Tag",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "deleted", DisplayValue: "deleted"},
									{Value: "file", DisplayValue: "file"},
									{Value: "folder", DisplayValue: "folder"},
								},
							},
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"server_modified": {
								DisplayName:  "Server Modified",
								ValueType:    "datetime",
								ProviderType: "datetime",
							},
						},
					},
					"file_requests": {
						DisplayName: "File Requests",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"is_open": {
								DisplayName:  "Is Open",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"shared_links": {
						DisplayName: "Shared Links",
						Fields: map[string]common.FieldMetadata{
							"url": {
								DisplayName:  "Url",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"shared_folders": {
						DisplayName: "Shared Folders",
						Fields: map[string]common.FieldMetadata{
							"shared_folder_id": {
								DisplayName:  "Shared Folder Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"mountable_shared_folders": {
						DisplayName: "Mountable Shared Folders",
						Fields: map[string]common.FieldMetadata{
							"shared_folder_id": {
								DisplayName:  "Shared Folder Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"received_files": {
						DisplayName: "Received Files",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"files", "search"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"files": {
						DisplayName: "Files",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
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
