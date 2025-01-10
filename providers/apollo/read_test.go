package apollo

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "empty.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	sequencesResponse := testutils.DataFromFile(t, "sequences.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "arsenal", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusBadRequest, string(unsupportedResponse)),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "opportunity_stages", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, string(zeroRecords)),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Sequences",
			Input: common.ReadParams{
				ObjectName: "sequences",
				Fields:     connectors.Fields("id", "name", "archived"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, string(sequencesResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":       "66e9e215ece19801b219997f",
						"name":     "Target Copywriting Clients in Dublin",
						"archived": false,
					},
					Raw: map[string]any{
						"id":                           "66e9e215ece19801b219997f",
						"name":                         "Target Copywriting Clients in Dublin",
						"archived":                     false,
						"created_at":                   "2024-09-17T20:09:57.837Z",
						"emailer_schedule_id":          "6095a711bd01d100a506d52a",
						"max_emails_per_day":           nil,
						"user_id":                      "66302798d03b9601c7934ebf",
						"same_account_reply_policy_cd": nil,
						"excluded_account_stage_ids": []any{
							"6095a710bd01d100a506d4b8",
							"6095a710bd01d100a506d4b9",
							"6095a710bd01d100a506d4ba",
							"6095a710bd01d100a506d4bb",
						},
						"excluded_contact_stage_ids": []any{
							"6095a710bd01d100a506d4b5",
							"6095a710bd01d100a506d4b4",
							"6095a710bd01d100a506d4b0",
							"6095a710bd01d100a506d4b1",
						},
						"contact_email_event_to_stage_mapping": map[string]any{},
						"label_ids": []any{
							"66e9e215ece19801b2199980",
							"66e9e215ece19801b2199981",
							"66e9e215ece19801b2199982",
						},
						"create_task_if_email_open":              false,
						"email_open_trigger_task_threshold":      float64(3),
						"mark_finished_if_click":                 false,
						"active":                                 false,
						"days_to_wait_before_mark_as_response":   float64(5),
						"starred_by_user_ids":                    []any{"66302798d03b9601c7934ebf"},
						"mark_finished_if_reply":                 true,
						"mark_finished_if_interested":            true,
						"mark_paused_if_ooo":                     true,
						"last_used_at":                           nil,
						"permissions":                            "team_can_use",
						"sequence_ruleset_id":                    "6095a711bd01d100a506d4e0",
						"folder_id":                              nil,
						"sequence_by_exact_daytime":              nil,
						"same_account_reply_delay_days":          float64(30),
						"is_performing_poorly":                   false,
						"num_contacts_email_status_extrapolated": float64(0),
						"remind_ab_test_results":                 false,
						"ab_test_step_ids":                       []any{},
						"prioritized_by_user":                    nil,
						"creation_type":                          "new",
						"num_steps":                              float64(3),
						"unique_scheduled":                       float64(0),
						"unique_delivered":                       float64(0),
						"unique_bounced":                         float64(0),
						"unique_opened":                          float64(0),
						"unique_hard_bounced":                    float64(0),
						"unique_spam_blocked":                    float64(0),
						"unique_replied":                         float64(0),
						"unique_demoed":                          float64(0),
						"unique_clicked":                         float64(0),
						"unique_unsubscribed":                    float64(0),
						"bounce_rate":                            float64(0),
						"hard_bounce_rate":                       float64(0),
						"open_rate":                              float64(0),
						"click_rate":                             float64(0),
						"reply_rate":                             float64(0),
						"spam_block_rate":                        float64(0),
						"opt_out_rate":                           float64(0),
						"demo_rate":                              float64(0),
						"loaded_stats":                           true,
						"cc_emails":                              "",
						"bcc_emails":                             "",
						"underperforming_touches_count":          float64(0),
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
