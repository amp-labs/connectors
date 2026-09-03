package monaco

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
			Name: "Metadata for every supported object",
			Input: []string{
				"accounts",
				"audiences",
				"campaigns",
				"contacts",
				"meetings",
				"opportunities",
				"sequenceTemplates",
				"sequences",
				"tags",
				"tasks",
				"users",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							// Nullable in the spec (anyOf [string, null]); the
							// generator collapses that to a plain string.
							"email": {
								DisplayName:  "email",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"do_not_contact": {
								DisplayName:  "do_not_contact",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
							// A genuine array stays "other" -- there is no
							// scalar type to report.
							"tags": {
								DisplayName:  "tags",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
							},
						},
					},
					"opportunities": {
						DisplayName: "Opportunities",
						Fields: map[string]common.FieldMetadata{
							"account_id": {
								DisplayName:  "account_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					// Served over GET rather than POST /list, and unpaginated.
					"tags": {
						DisplayName: "Tags",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					"sequenceTemplates": {
						DisplayName: "Sequence Templates",
						Fields: map[string]common.FieldMetadata{
							"is_default": {
								DisplayName:  "is_default",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:       "Unknown object is reported, known objects still resolve",
			Input:      []string{"contacts", "butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"butterflies": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
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
