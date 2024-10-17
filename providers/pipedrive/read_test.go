package pipedrive

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	nextPageTest := testutils.DataFromFile(t, "activities.json")
	leads := testutils.DataFromFile(t, "leads.json")

	ErrResponseBody := `{
		"success":false,
		"error":"Scope and URL mismatch",
		"errorCode":403,
		"error_info":"Please check developers.pipedrive.com"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(nextPageTest)
	}))

	tests := []testroutines.Read{
		{
			Name:         "Object Name Required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At Least One Read Field Required",
			Input:        common.ReadParams{ObjectName: "activities"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "NextPage URL construction",
			Input:        common.ReadParams{ObjectName: "activities", Fields: connectors.Fields("id")},
			Server:       server,
			ExpectedErrs: nil,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": float64(2),
						},
						Raw: map[string]any{
							"id":                            float64(2),
							"company_id":                    float64(13313052),
							"user_id":                       float64(20580207),
							"done":                          false,
							"type":                          "call",
							"reference_type":                nil,
							"reference_id":                  nil,
							"conference_meeting_client":     nil,
							"conference_meeting_url":        nil,
							"due_date":                      "2024-10-30",
							"due_time":                      "",
							"duration":                      "",
							"busy_flag":                     false,
							"add_time":                      "2024-10-16 12:16:02",
							"marked_as_done_time":           "",
							"last_notification_time":        nil,
							"last_notification_user_id":     nil,
							"notification_language_id":      nil,
							"subject":                       "I usually can't come up with words",
							"public_description":            "Demo activity",
							"calendar_sync_include_context": nil,
							"location":                      "Dar es salaam",
							"org_id":                        nil,
							"person_id":                     nil,
							"deal_id":                       nil,
							"lead_id":                       nil,
							"active_flag":                   true,
							"update_time":                   "2024-10-16 12:16:02",
							"update_user_id":                nil,
							"source_timezone":               nil,
							"rec_rule":                      nil,
							"rec_rule_extension":            nil,
							"rec_master_activity_id":        nil,
							"conference_meeting_id":         nil,
							"original_start_time":           nil,
							"private":                       false,
							"priority":                      nil,
							"note":                          nil,
							"created_by_user_id":            float64(20580207),
							"location_subpremise":           nil,
							"location_street_number":        nil,
							"location_route":                nil,
							"location_sublocality":          nil,
							"location_locality":             nil,
							"location_admin_area_level_1":   nil,
							"location_admin_area_level_2":   nil,
							"location_country":              nil,
							"location_postal_code":          nil,
							"location_formatted_address":    nil,
							"attendees":                     nil,
							"participants":                  nil,
							"series":                        nil,
							"is_recurring":                  nil,
							"org_name":                      nil,
							"person_name":                   nil,
							"deal_title":                    nil,
							"lead_title":                    nil,
							"owner_name":                    "Integration User",
							"person_dropbox_bcc":            nil,
							"deal_dropbox_bcc":              nil,
							"assigned_to_user_id":           float64(20580207),
							"type_name":                     "Call",
							"lead":                          nil,
						},
					},
				},
				Done:     false,
				NextPage: common.NextPageToken(fmt.Sprintf("%s/v1/activities?start=1", server.URL)),
			},
		},
		{
			Name:  "Not Found Resource or Higher Suite Resource",
			Input: common.ReadParams{ObjectName: "activitiess", Fields: connectors.Fields("id")},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				mockutils.WriteBody(w, ErrResponseBody)
			})),
			ExpectedErrs: []error{common.NewHTTPStatusError(http.StatusForbidden,
				fmt.Errorf("%w: %s", common.ErrForbidden, ErrResponseBody))},
			Expected: nil,
		},
		{
			Name: "Successful Read",
			Input: common.ReadParams{
				ObjectName: "leads",
				Fields:     connectors.Fields("channel", "id", "origin", "title"),
			},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(leads)
			})),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				// custom comparison focuses on subset of fields to keep the test short
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"channel": nil,
						"id":      "8ba414d0-8ad8-11ef-9e9e-07879dd74146",
						"origin":  "ManuallyCreated",
						"title":   "Ampersand lead",
					},
					Raw: map[string]any{
						"add_time":            "2024-10-15T09:33:33.213Z",
						"cc_email":            "integrationuser-sandbox2+13313052+leadif7ZKQ7mhWz28LivD9BmE3@pipedrivemail.com",
						"channel":             nil,
						"channel_id":          nil,
						"creator_id":          float64(20580207),
						"expected_close_date": nil,
						"id":                  "8ba414d0-8ad8-11ef-9e9e-07879dd74146",
						"is_archived":         false,
						"label_ids":           []any{},
						"next_activity_id":    nil,
						"organization_id":     float64(2),
						"origin":              "ManuallyCreated",
						"origin_id":           nil,
						"owner_id":            float64(20580207),
						"person_id":           float64(2),
						"source_name":         "Manually created",
						"title":               "Ampersand lead",
						"update_time":         "2024-10-15T09:33:33.213Z",
						"value":               nil,
						"visible_to":          "3",
						"was_seen":            true,
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
