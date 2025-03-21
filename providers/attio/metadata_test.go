// nolint
package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	listresponse := testutils.DataFromFile(t, "lists.json")
	notesresponse := testutils.DataFromFile(t, "notes.json")
	workspacemembersresponse := testutils.DataFromFile(t, "workspace_members.json")
	tasksresponse := testutils.DataFromFile(t, "tasks.json")
	companiesresponse := testutils.DataFromFile(t, "companies.json")
	companiesObjectResponse := []byte(`{"data": {"plural_noun": "Companies"}}`)

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"lists", "workspace_members", "notes", "tasks", "companies"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/v2/lists"),
					Then: mockserver.Response(http.StatusOK, listresponse),
				}, {
					If:   mockcond.PathSuffix("/v2/workspace_members"),
					Then: mockserver.Response(http.StatusOK, workspacemembersresponse),
				}, {
					If:   mockcond.PathSuffix("/v2/notes"),
					Then: mockserver.Response(http.StatusOK, notesresponse),
				}, {
					If:   mockcond.PathSuffix("/v2/tasks"),
					Then: mockserver.Response(http.StatusOK, tasksresponse),
				}, {
					If:   mockcond.PathSuffix("/v2/objects/companies/attributes"),
					Then: mockserver.Response(http.StatusOK, companiesresponse),
				}, {
					If:   mockcond.PathSuffix("/v2/objects/companies"),
					Then: mockserver.Response(http.StatusOK, companiesObjectResponse),
				},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lists": {
						DisplayName: "lists",
						Fields: map[string]common.FieldMetadata{
							"api_slug": {
								DisplayName:  "api_slug",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"api_slug": "api_slug",
							"id":       "id",
							"name":     "name",
						},
					},
					"workspace_members": {
						DisplayName: "workspace_members",
						Fields: map[string]common.FieldMetadata{
							"email_address": {
								DisplayName:  "email_address",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"last_name": {
								DisplayName:  "last_name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"email_address": "email_address",
							"first_name":    "first_name",
							"id":            "id",
							"last_name":     "last_name",
						},
					},
					"notes": {
						DisplayName: "notes",
						Fields: map[string]common.FieldMetadata{
							"content_plaintext": {
								DisplayName:  "content_plaintext",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"content_plaintext": "content_plaintext",
							"id":                "id",
							"title":             "title",
						},
					},
					"tasks": {
						DisplayName: "tasks",
						Fields: map[string]common.FieldMetadata{
							"content_plaintext": {
								DisplayName:  "content_plaintext",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"content_plaintext": "content_plaintext",
							"id":                "id",
						},
					},
					"companies": {
						DisplayName: "Companies",
						Fields: map[string]common.FieldMetadata{
							"record_id": {
								DisplayName:  "record_id",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     true,
								Values:       nil,
							},
							"domains": {
								DisplayName:  "domains",
								ValueType:    "other",
								ProviderType: "domain",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "datetime",
								ProviderType: "timestamp",
								ReadOnly:     true,
								Values:       nil,
							},
							"created_by": {
								DisplayName:  "created_by",
								ValueType:    "other",
								ProviderType: "actor-reference",
								ReadOnly:     true,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"record_id":  "record_id",
							"domains":    "domains",
							"name":       "name",
							"created_at": "created_at",
							"created_by": "created_by",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)

	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(serverURL)

	return connector, nil
}
