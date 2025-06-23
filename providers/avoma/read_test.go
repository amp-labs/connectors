package avoma

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseMeetings := testutils.DataFromFile(t, "meetings.json")
	responseUsers := testutils.DataFromFile(t, "users.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read list of all meetings",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields(""),
				Since:      time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2025, time.June, 3, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseMeetings),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"uuid":               "bb4097e2-0e18-43a6-845d-49563f62ff15",
						"subject":            "Meeting with Integration",
						"start_at":           "2025-06-03T23:00:00Z",
						"end_at":             "2025-06-03T23:30:00Z",
						"created":            "2025-06-02T09:44:13.064654Z",
						"modified":           "2025-06-03T23:15:42.405303Z",
						"is_internal":        false,
						"is_private":         false,
						"is_call":            false,
						"organizer_email":    "integration.test@withampersand.com",
						"type":               nil,
						"outcome":            nil,
						"transcription_uuid": nil,
						"attendees": []any{
							map[string]any{
								"email":           "integration.test@withampersand.com",
								"name":            "Integration Test",
								"uuid":            "74d13b27-287f-4e85-a3a5-77f3734b5197",
								"response_status": "accepted",
							},
							map[string]any{
								"email":           "josephbkarage@gmail.com",
								"name":            "Josephbkarage",
								"uuid":            "4b317326-93d3-498b-94ee-f5acfec05bca",
								"response_status": "accepted",
							},
						},
						"duration":          nil,
						"state":             "in_progress",
						"call_details":      nil,
						"audio_ready":       false,
						"video_ready":       false,
						"transcript_ready":  false,
						"notes_ready":       false,
						"processing_status": "recording_error",
						"url":               "https://app.avoma.com/meetings/bb4097e2-0e18-43a6-845d-49563f62ff15",
						"recording_uuid":    nil,
						"purpose":           nil,
					},
				},
				},
				NextPage: "https://api.avoma.com/v1/meetings?from_date=2025-06-01T00:00:00Z&page_size=50&to_date=2025-06-03T00:00:00Z",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all users",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"uuid": "bbe512ba-5eb8-48c7-ab1f-9c069836436a",
						"user": map[string]any{
							"email":        "beadko@gmail.com",
							"first_name":   "Beatrise",
							"last_name":    "",
							"profile_pic":  "https://lh3.googleusercontent.com/a/ACg8ocJIua6Ibz2UjwZN2QJrouKsGokM0b8-myrfCzhZLMuBP_7kRe6UIw=s96-c",
							"position":     "developer",
							"is_active":    false,
							"job_function": "Engineering",
							"upn":          nil,
						},
						"role": map[string]any{
							"name":         "member",
							"display_name": "member",
							"description":  "Full access to your meetings. Can listen to meetings of other members from the same team",
							"uuid":         "051b16a4-039d-4422-97f6-2d7337542741",
							"role_type":    "sys",
						},
						"position": "developer",
						"teams":    []any{},
						"status":   "active",
						"active":   false,
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
