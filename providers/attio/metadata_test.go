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
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_by_actor": {
								DisplayName:  "created_by_actor",
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
							"parent_object": {
								DisplayName:  "parent_object",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"workspace_access": {
								DisplayName:  "workspace_access",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"workspace_member_access": {
								DisplayName:  "workspace_member_access",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
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
						Fields: map[string]common.FieldMetadata{
							"access_level": {
								DisplayName:  "access_level",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"avatar_url": {
								DisplayName:  "avatar_url",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
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
						Fields: map[string]common.FieldMetadata{
							"content_plaintext": {
								DisplayName:  "content_plaintext",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_by_actor": {
								DisplayName:  "created_by_actor",
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
							"parent_object": {
								DisplayName:  "parent_object",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"parent_record_id": {
								DisplayName:  "parent_record_id",
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
						Fields: map[string]common.FieldMetadata{
							"assignees": {
								DisplayName:  "assignees",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"content_plaintext": {
								DisplayName:  "content_plaintext",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_by_actor": {
								DisplayName:  "created_by_actor",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"deadline_at": {
								DisplayName:  "deadline_at",
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
							"is_completed": {
								DisplayName:  "is_completed",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"linked_records": {
								DisplayName:  "linked_records",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
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
						Fields: map[string]common.FieldMetadata{
							"record_id": {
								DisplayName:  "record_id",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
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
							"description": {
								DisplayName:  "description",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"team": {
								DisplayName:  "team",
								ValueType:    "other",
								ProviderType: "record-reference",
								ReadOnly:     false,
								Values:       nil,
							},
							"categories": {
								DisplayName:  "categories",
								ValueType:    "singleSelect",
								ProviderType: "select",
								ReadOnly:     false,
								Values:       nil,
							},
							"primary_location": {
								DisplayName:  "primary_location",
								ValueType:    "other",
								ProviderType: "location",
								ReadOnly:     false,
								Values:       nil,
							},
							"logo_url": {
								DisplayName:  "logo_url",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"angellist": {
								DisplayName:  "angellist",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"facebook": {
								DisplayName:  "facebook",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"instagram": {
								DisplayName:  "instagram",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"linkedin": {
								DisplayName:  "linkedin",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"twitter": {
								DisplayName:  "twitter",
								ValueType:    "string",
								ProviderType: "text",
								ReadOnly:     false,
								Values:       nil,
							},
							"twitter_follower_count": {
								DisplayName:  "twitter_follower_count",
								ValueType:    "int",
								ProviderType: "number",
								ReadOnly:     false,
								Values:       nil,
							},
							"estimated_arr_usd": {
								DisplayName:  "estimated_arr_usd",
								ValueType:    "singleSelect",
								ProviderType: "select",
								ReadOnly:     false,
								Values:       nil,
							},
							"funding_raised_usd": {
								DisplayName:  "funding_raised_usd",
								ValueType:    "other",
								ProviderType: "currency",
								ReadOnly:     false,
								Values:       nil,
							},
							"foundation_date": {
								DisplayName:  "foundation_date",
								ValueType:    "date",
								ProviderType: "date",
								ReadOnly:     false,
								Values:       nil,
							},
							"employee_range": {
								DisplayName:  "employee_range",
								ValueType:    "singleSelect",
								ProviderType: "select",
								ReadOnly:     false,
								Values:       nil,
							},
							"first_calendar_interaction": {
								DisplayName:  "first_calendar_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"last_calendar_interaction": {
								DisplayName:  "last_calendar_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"next_calendar_interaction": {
								DisplayName:  "next_calendar_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"first_email_interaction": {
								DisplayName:  "first_email_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"last_email_interaction": {
								DisplayName:  "last_email_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"first_call_interaction": {
								DisplayName:  "first_call_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"last_call_interaction": {
								DisplayName:  "last_call_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"next_call_interaction": {
								DisplayName:  "next_call_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"first_in_person_meeting_interaction": {
								DisplayName:  "first_in_person_meeting_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"last_in_person_meeting_interaction": {
								DisplayName:  "last_in_person_meeting_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"next_in_person_meeting_interaction": {
								DisplayName:  "next_in_person_meeting_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"first_interaction": {
								DisplayName:  "first_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"last_interaction": {
								DisplayName:  "last_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"next_interaction": {
								DisplayName:  "next_interaction",
								ValueType:    "other",
								ProviderType: "interaction",
								ReadOnly:     false,
								Values:       nil,
							},
							"strongest_connection_strength_legacy": {
								DisplayName:  "strongest_connection_strength_legacy",
								ValueType:    "int",
								ProviderType: "number",
								ReadOnly:     false,
								Values:       nil,
							},
							"strongest_connection_strength": {
								DisplayName:  "strongest_connection_strength",
								ValueType:    "singleSelect",
								ProviderType: "select",
								ReadOnly:     false,
								Values:       nil,
							},
							"strongest_connection_user": {
								DisplayName:  "strongest_connection_user",
								ValueType:    "other",
								ProviderType: "actor-reference",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "datetime",
								ProviderType: "timestamp",
								ReadOnly:     false,
								Values:       nil,
							},
							"created_by": {
								DisplayName:  "created_by",
								ValueType:    "other",
								ProviderType: "actor-reference",
								ReadOnly:     false,
								Values:       nil,
							},
						},
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
