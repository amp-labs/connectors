// nolint
package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	objectresponse := testutils.DataFromFile(t, "objects.json")
	listresponse := testutils.DataFromFile(t, "lists.json")
	notesresponse := testutils.DataFromFile(t, "notes.json")
	workspacemembersresponse := testutils.DataFromFile(t, "workspace_members.json")
	webhooksresponse := testutils.DataFromFile(t, "webhooks.json")
	tasksresponse := testutils.DataFromFile(t, "tasks.json")

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"objects", "lists", "workspace_members", "notes", "webhooks", "tasks"},
			Server: mockserver.Crossroad{
				Setup: mockserver.ContentJSON(),
				Paths: []mockserver.Path{{
					Condition: mockcond.PathSuffix("/v2/objects"),
					OnSuccess: mockserver.Response(http.StatusOK, objectresponse),
				}, {
					Condition: mockcond.PathSuffix("/v2/lists"),
					OnSuccess: mockserver.Response(http.StatusOK, listresponse),
				}, {
					Condition: mockcond.PathSuffix("/v2/workspace_members"),
					OnSuccess: mockserver.Response(http.StatusOK, workspacemembersresponse),
				}, {
					Condition: mockcond.PathSuffix("/v2/notes"),
					OnSuccess: mockserver.Response(http.StatusOK, notesresponse),
				}, {
					Condition: mockcond.PathSuffix("/v2/tasks"),
					OnSuccess: mockserver.Response(http.StatusOK, tasksresponse),
				}, {
					Condition: mockcond.PathSuffix("/v2/webhooks"),
					OnSuccess: mockserver.Response(http.StatusOK, webhooksresponse),
				}},
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"objects": {
						DisplayName: "objects",
						FieldsMap: map[string]string{
							"api_slug":      "api_slug",
							"created_at":    "created_at",
							"id":            "id",
							"plural_noun":   "plural_noun",
							"singular_noun": "singular_noun",
						},
					},
					"lists": {
						DisplayName: "lists",
						FieldsMap: map[string]string{
							"api_slug":                "api_slug",
							"created_at":              "created_at",
							"created_by_actor":        "created_by_actor",
							"id":                      "id",
							"name":                    "name",
							"parent_object":           "parent_object",
							"workspace_access":        "workspace_access",
							"workspace_member_access": "workspace_member_access",
						},
					},
					"workspace_members": {
						DisplayName: "workspace_members",
						FieldsMap: map[string]string{
							"access_level":  "access_level",
							"avatar_url":    "avatar_url",
							"created_at":    "created_at",
							"email_address": "email_address",
							"first_name":    "first_name",
							"id":            "id",
							"last_name":     "last_name",
						},
					},
					"webhooks": {
						DisplayName: "webhooks",
						FieldsMap: map[string]string{
							"created_at":    "created_at",
							"id":            "id",
							"status":        "status",
							"subscriptions": "subscriptions",
							"target_url":    "target_url",
						},
					},
					"notes": {
						DisplayName: "notes",
						FieldsMap: map[string]string{
							"content_plaintext": "content_plaintext",
							"created_at":        "created_at",
							"created_by_actor":  "created_by_actor",
							"id":                "id",
							"parent_object":     "parent_object",
							"parent_record_id":  "parent_record_id",
							"title":             "title",
						},
					},
					"tasks": {
						DisplayName: "tasks",
						FieldsMap: map[string]string{
							"assignees":         "assignees",
							"content_plaintext": "content_plaintext",
							"created_at":        "created_at",
							"created_by_actor":  "created_by_actor",
							"deadline_at":       "deadline_at",
							"id":                "id",
							"is_completed":      "is_completed",
							"linked_records":    "linked_records",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine.
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
