// nolint
package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	listResponse := testutils.DataFromFile(t, "lists.json")
	notesResponse := testutils.DataFromFile(t, "notes.json")
	workspacemembersResponse := testutils.DataFromFile(t, "workspace_members.json")
	tasksResponse := testutils.DataFromFile(t, "tasks.json")
	companiesResponse := testutils.DataFromFile(t, "companies.json")
	optionsTeamAttributeResponse := testutils.DataFromFile(t, "option_team_attribute.json")
	companiesObjectResponse := []byte(`{"data": {"plural_noun": "Companies"}}`)
	usersResponse := testutils.DataFromFile(t, "users.json")
	optionsResponse := testutils.DataFromFile(t, "options.json")
	usersObjectResponse := []byte(`{"data": {"plural_noun": "Users"}}`)

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"lists", "workspace_members", "notes", "tasks", "companies", "users"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/v2/lists"),
						Then: mockserver.Response(http.StatusOK, listResponse),
					}, {
						If:   mockcond.Path("/v2/workspace_members"),
						Then: mockserver.Response(http.StatusOK, workspacemembersResponse),
					}, {
						If:   mockcond.Path("/v2/notes"),
						Then: mockserver.Response(http.StatusOK, notesResponse),
					}, {
						If:   mockcond.Path("/v2/tasks"),
						Then: mockserver.Response(http.StatusOK, tasksResponse),
					}, {
						If:   mockcond.Path("/v2/objects/companies/attributes"),
						Then: mockserver.Response(http.StatusOK, companiesResponse),
					}, {
						If:   mockcond.Path("/v2/objects/companies"),
						Then: mockserver.Response(http.StatusOK, companiesObjectResponse),
					}, {
						If:   mockcond.Path("/v2/objects/users/attributes"),
						Then: mockserver.Response(http.StatusOK, usersResponse),
					}, {
						If:   mockcond.Path("/v2/objects/users"),
						Then: mockserver.Response(http.StatusOK, usersObjectResponse),
					}, {
						If:   mockcond.Path("/v2/objects/ffbca575-69c4-4080-bf98-91d79aeea4b1/attributes/89c07285-4d31-4fa7-9cbf-779c5f4debf1/options"),
						Then: mockserver.Response(http.StatusOK, optionsResponse),
					}, {
						If:   mockcond.Path("/v2/objects/1a4b88cf-520e-4394-886b-941c07c78854/records/query"),
						Then: mockserver.Response(http.StatusOK, optionsTeamAttributeResponse),
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
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
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
								Values:       nil,
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"last_name": {
								DisplayName:  "last_name",
								ValueType:    "other",
								ProviderType: "",
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
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "other",
								ProviderType: "",
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
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
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
								ReadOnly:     goutils.Pointer(true),
								Values:       nil,
							},
							"domains": {
								DisplayName:  "domains",
								ValueType:    "multiSelect",
								ProviderType: "domain",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"team": {
								DisplayName:  "team",
								ValueType:    "multiSelect",
								ProviderType: "record-reference",
								ReadOnly:     goutils.Pointer(false),
								Values: common.FieldValues{
									{
										Value:        "d0be3734-3b4d-4094-9925-9dd906941197",
										DisplayValue: "d0be3734-3b4d-4094-9925-9dd906941197",
									},
								},
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "datetime",
								ProviderType: "timestamp",
								ReadOnly:     goutils.Pointer(true),
								Values:       nil,
							},
							"created_by": {
								DisplayName:  "created_by",
								ValueType:    "other",
								ProviderType: "actor-reference",
								ReadOnly:     goutils.Pointer(true),
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
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"record_id": {
								DisplayName:  "record_id",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     goutils.Pointer(true),
								Values:       nil,
							},
							"user_id": {
								DisplayName:  "user_id",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     goutils.Pointer(false),
								Values:       nil,
							},
							"education": {
								DisplayName:  "education",
								ValueType:    "multiSelect",
								ProviderType: "select",
								ReadOnly:     goutils.Pointer(false),
								Values: common.FieldValues{
									{Value: "UG", DisplayValue: "UG"},
									{Value: "PG", DisplayValue: "PG"},
									{Value: "Diploma", DisplayValue: "Diploma"},
								},
							},
						},
						FieldsMap: map[string]string{
							"record_id": "record_id",
							"user_id":   "user_id",
							"education": "education",
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
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
