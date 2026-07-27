package sendgrid

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
			Name:       "Successful metadata for email objects",
			Input:      []string{"lists", "templates", "bounces"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lists": {
						DisplayName: "Lists",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "List Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"templates": {
						DisplayName: "Templates",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Template Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"bounces": {
						DisplayName: "Bounces",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "Created",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
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
			Name:         "Empty objects returns missing objects error",
			Input:        nil,
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"lists", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lists": {
						DisplayName: "Lists",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "List Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"unknown_object": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
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

	connector.SetUnitTestMockServerBaseURL(serverURL)

	return connector, nil
}
