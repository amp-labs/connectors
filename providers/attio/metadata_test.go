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
					"companies": {
						DisplayName: "Companies",
						FieldsMap: map[string]string{
							"record_id":                            "record_id",
							"domains":                              "domains",
							"name":                                 "name",
							"description":                          "description",
							"team":                                 "team",
							"categories":                           "categories",
							"primary_location":                     "primary_location",
							"logo_url":                             "logo_url",
							"angellist":                            "angellist",
							"facebook":                             "facebook",
							"instagram":                            "instagram",
							"linkedin":                             "linkedin",
							"twitter":                              "twitter",
							"twitter_follower_count":               "twitter_follower_count",
							"estimated_arr_usd":                    "estimated_arr_usd",
							"funding_raised_usd":                   "funding_raised_usd",
							"foundation_date":                      "foundation_date",
							"employee_range":                       "employee_range",
							"first_calendar_interaction":           "first_calendar_interaction",
							"last_calendar_interaction":            "last_calendar_interaction",
							"next_calendar_interaction":            "next_calendar_interaction",
							"first_email_interaction":              "first_email_interaction",
							"last_email_interaction":               "last_email_interaction",
							"first_call_interaction":               "first_call_interaction",
							"last_call_interaction":                "last_call_interaction",
							"next_call_interaction":                "next_call_interaction",
							"first_in_person_meeting_interaction":  "first_in_person_meeting_interaction",
							"last_in_person_meeting_interaction":   "last_in_person_meeting_interaction",
							"next_in_person_meeting_interaction":   "next_in_person_meeting_interaction",
							"first_interaction":                    "first_interaction",
							"last_interaction":                     "last_interaction",
							"next_interaction":                     "next_interaction",
							"strongest_connection_strength_legacy": "strongest_connection_strength_legacy",
							"strongest_connection_strength":        "strongest_connection_strength",
							"strongest_connection_user":            "strongest_connection_user",
							"created_at":                           "created_at",
							"created_by":                           "created_by",
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
