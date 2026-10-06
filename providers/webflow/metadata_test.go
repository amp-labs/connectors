package webflow

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
			Name: "Successful metadata for account and site scoped objects",
			Input: []string{
				"sites",
				"collections",
				"pages",
				"form_submissions",
				"orders",
				"webhooks",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"sites": {
						DisplayName: "Sites",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"displayName": {
								DisplayName:  "displayName",
								ValueType:    "string",
								ProviderType: "string",
							},
							"dataCollectionType": {
								DisplayName:  "dataCollectionType",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "always", DisplayValue: "always"},
									{Value: "optOut", DisplayValue: "optOut"},
									{Value: "disabled", DisplayValue: "disabled"},
								},
							},
						},
					},
					"collections": {
						DisplayName: "Collections",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"singularName": {
								DisplayName:  "singularName",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"pages": {
						DisplayName: "Pages",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"archived": {
								DisplayName:  "archived",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"form_submissions": {
						DisplayName: "Form Submissions",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"formResponse": {
								DisplayName:  "formResponse",
								ValueType:    "other",
								ProviderType: "object",
							},
						},
					},
					"orders": {
						DisplayName: "Orders",
						Fields: map[string]common.FieldMetadata{
							"orderId": {
								DisplayName:  "orderId",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"webhooks": {
						DisplayName: "Webhooks",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"url": {
								DisplayName:  "url",
								ValueType:    "string",
								ProviderType: "string",
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
			Input:      []string{"sites", "items"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"sites": {
						DisplayName: "Sites",
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
					"items": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
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
