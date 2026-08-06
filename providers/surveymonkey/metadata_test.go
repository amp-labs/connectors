package surveymonkey

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
			Name: "Successful metadata for core objects",
			Input: []string{
				"groups",
				"survey_categories",
				"survey_templates",
				"team_survey_templates",
				"survey_languages",
				"question_bank_questions",
				"survey_folders",
				"contacts",
				"contact_lists",
				"organizations",
				"roles",
				"benchmark_bundles",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"groups": {
						DisplayName: "Groups",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Group Id",
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
					"survey_categories": {
						DisplayName: "Survey Categories",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Category Id",
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
					"survey_templates": {
						DisplayName: "Survey Templates",
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
					"team_survey_templates": {
						DisplayName: "Team Survey Templates",
						Fields: map[string]common.FieldMetadata{
							"team_template_id": {
								DisplayName:  "Team Template Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"survey_id": {
								DisplayName:  "Survey Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"survey_languages": {
						DisplayName: "Survey Languages",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Language Code",
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
					"question_bank_questions": {
						DisplayName: "Question Bank Questions",
						Fields: map[string]common.FieldMetadata{
							"question_id": {
								DisplayName:  "Question Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"text": {
								DisplayName:  "Text",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"survey_folders": {
						DisplayName: "Survey Folders",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Folder Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "Title",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Contact Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"contact_lists": {
						DisplayName: "Contact Lists",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Contact List Id",
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
					"organizations": {
						DisplayName: "Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Organization Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"roles": {
						DisplayName: "Roles",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Role Id",
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
					"benchmark_bundles": {
						DisplayName: "Benchmark Bundles",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Benchmark Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "Title",
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
			Input:      []string{"groups", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"groups": {
						DisplayName: "Groups",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Group Id",
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
