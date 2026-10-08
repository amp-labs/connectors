package wrike

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
			Name: "Successful metadata for core work management objects",
			Input: []string{
				"tasks",
				"folders",
				"contacts",
				"customfields",
				"comments",
				"timelogs",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tasks": {
						DisplayName: "Tasks",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
							},
							"importance": {
								DisplayName:  "importance",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "High", DisplayValue: "High"},
									{Value: "Low", DisplayValue: "Low"},
									{Value: "Normal", DisplayValue: "Normal"},
								},
							},
						},
					},
					"folders": {
						DisplayName: "Folders & Projects",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "string",
								ProviderType: "string",
							},
							"deleted": {
								DisplayName:  "deleted",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"customfields": {
						DisplayName: "Custom Fields",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"comments": {
						DisplayName: "Comments",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"text": {
								DisplayName:  "text",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"timelogs": {
						DisplayName: "Timelogs",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"hours": {
								DisplayName:  "hours",
								ValueType:    "other",
								ProviderType: "number",
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
			Input:      []string{"tasks", "account"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tasks": {
						DisplayName: "Tasks",
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
					"account": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
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
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
