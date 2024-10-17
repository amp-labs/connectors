package pipedrive

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "zero-records.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "A success API Response",
			Input: []string{"currencies", "filters"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(zeroRecords))
			})),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"currencies": {
						DisplayName: "Currencies",
						FieldsMap: map[string]string{
							"active_flag":    "active_flag",
							"code":           "code",
							"decimal_points": "decimal_points",
							"id":             "id",
							"is_custom_flag": "is_custom_flag",
							"name":           "name",
							"symbol":         "symbol",
						},
					},
					"filters": {
						DisplayName: "Filters",
						FieldsMap: map[string]string{
							"active_flag":    "active_flag",
							"add_time":       "add_time",
							"custom_view_id": "custom_view_id",
							"id":             "id",
							"name":           "name",
							"type":           "type",
							"update_time":    "update_time",
							"user_id":        "user_id",
							"visible_to":     "visible_to",
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Zero records returned from server",
			Input: []string{"activities"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(zeroRecords))
			})),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities": {
						DisplayName: "Activities",
						FieldsMap: map[string]string{
							"active_flag":                   "active_flag",
							"add_time":                      "add_time",
							"assigned_to_user_id":           "assigned_to_user_id",
							"attendees":                     "attendees",
							"busy_flag":                     "busy_flag",
							"calendar_sync_include_context": "calendar_sync_include_context",
							"company_id":                    "company_id",
							"conference_meeting_client":     "conference_meeting_client",
							"conference_meeting_id":         "conference_meeting_id",
							"conference_meeting_url":        "conference_meeting_url",
							"created_by_user_id":            "created_by_user_id",
							"deal_dropbox_bcc":              "deal_dropbox_bcc",
							"deal_id":                       "deal_id",
							"deal_title":                    "deal_title",
							"done":                          "done",
							"due_date":                      "due_date",
							"due_time":                      "due_time",
							"duration":                      "duration",
							"file":                          "file",
							"gcal_event_id":                 "gcal_event_id",
							"google_calendar_etag":          "google_calendar_etag",
							"google_calendar_id":            "google_calendar_id",
							"id":                            "id",
							"last_notification_time":        "last_notification_time",
							"last_notification_user_id":     "last_notification_user_id",
							"lead_id":                       "lead_id",
							"location":                      "location",
							"location_admin_area_level_1":   "location_admin_area_level_1",
							"location_admin_area_level_2":   "location_admin_area_level_2",
							"location_country":              "location_country",
							"location_formatted_address":    "location_formatted_address",
							"location_locality":             "location_locality",
							"location_postal_code":          "location_postal_code",
							"location_route":                "location_route",
							"location_street_number":        "location_street_number",
							"location_sublocality":          "location_sublocality",
							"location_subpremise":           "location_subpremise",
							"marked_as_done_time":           "marked_as_done_time",
							"note":                          "note",
							"notification_language_id":      "notification_language_id",
							"org_id":                        "org_id",
							"org_name":                      "org_name",
							"owner_name":                    "owner_name",
							"participants":                  "participants",
							"person_dropbox_bcc":            "person_dropbox_bcc",
							"person_id":                     "person_id",
							"person_name":                   "person_name",
							"project_id":                    "project_id",
							"public_description":            "public_description",
							"rec_master_activity_id":        "rec_master_activity_id",
							"rec_rule":                      "rec_rule",
							"rec_rule_extension":            "rec_rule_extension",
							"reference_id":                  "reference_id",
							"reference_type":                "reference_type",
							"series":                        "series",
							"source_timezone":               "source_timezone",
							"subject":                       "subject",
							"type":                          "type",
							"update_time":                   "update_time",
							"update_user_id":                "update_user_id",
							"user_id":                       "user_id",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(WithAuthenticatedClient(http.DefaultClient))
	if err != nil {
		return nil, err
	}

	connector.setBaseURL(serverURL)

	return connector, nil
}
