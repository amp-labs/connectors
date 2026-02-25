package salesloft

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordByIds(t *testing.T) {
	t.Parallel()

	responseGetPeople := testutils.DataFromFile(t, "get-records-people.json")
	responseGetAccounts := testutils.DataFromFile(t, "get-records-accounts.json")
	responseGetUsers := testutils.DataFromFile(t, "read-list-users.json")

	tests := []testroutines.TestCase[common.ReadByIdsParams, []common.ReadResultRow]{
		{
			Name:         "Missing object name returns error",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully fetch people by IDs",
			Input: common.ReadByIdsParams{
				ObjectName: "people",
				Fields:     []string{"email_address"},
				RecordIds:       []string{"164510523", "164510464"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/people"),
					mockcond.QueryParam("ids[]", "164510523", "164510464"),
				},
				Then: mockserver.Response(http.StatusOK, responseGetPeople),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "164510523",
					Fields: map[string]any{
						"id":            float64(164510523),
						"email_address": "losbourn29@paypal.com",
					},
					Raw: map[string]any{
						"id":             float64(164510523),
						"first_name":     "Lynnelle",
						"last_name":      "new",
						"email_address":  "losbourn29@paypal.com",
						"do_not_contact": false,
						"custom_fields":  map[string]any{},
					},
				},
				{
					Id: "164510464",
					Fields: map[string]any{
						"id":            float64(164510464),
						"email_address": "losbourn27@paypal.com",
					},
					Raw: map[string]any{
						"id":             float64(164510464),
						"first_name":     "Lynnelle",
						"last_name":      "Osbourn",
						"email_address":  "losbourn27@paypal.com",
						"do_not_contact": false,
						"custom_fields":  map[string]any{},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully fetch accounts by IDs",
			Input: common.ReadByIdsParams{
				ObjectName: "accounts",
				Fields:     []string{"name"},
				RecordIds:       []string{"48371814", "48371806"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/accounts"),
					mockcond.QueryParam("ids[]", "48371814", "48371806"),
				},
				Then: mockserver.Response(http.StatusOK, responseGetAccounts),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "48371814",
					Fields: map[string]any{
						"id":   float64(48371814),
						"name": "Reebok",
					},
					Raw: map[string]any{
						"id":             float64(48371814),
						"name":           "Reebok",
						"domain":         "https://www.reebok.eu/en-gb/",
						"do_not_contact": false,
					},
				},
				{
					Id: "48371806",
					Fields: map[string]any{
						"id":   float64(48371806),
						"name": "Asics",
					},
					Raw: map[string]any{
						"id":             float64(48371806),
						"name":           "Asics",
						"domain":         "https://www.asics.com/gb/en-gb/",
						"do_not_contact": false,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully fetch users by IDs",
			Input: common.ReadByIdsParams{
				ObjectName: "users",
				Fields:     []string{"email"},
				RecordIds:       []string{"49067"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/users"),
					mockcond.QueryParam("ids[]", "49067"),
				},
				Then: mockserver.Response(http.StatusOK, responseGetUsers),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "49067",
					Fields: map[string]any{
						"id":    float64(49067),
						"email": "somebody@withampersand.com",
					},
					Raw: map[string]any{
						"id":                         float64(49067),
						"guid":                       "0863ed13-7120-479b-8650-206a3679e2fb",
						"created_at":                 "2024-02-20T20:03:20.996772-05:00",
						"updated_at":                 "2024-04-27T07:18:23.531659-04:00",
						"name":                       "Int User",
						"first_name":                 "Int",
						"last_name":                  "User",
						"job_role":                   nil,
						"active":                     true,
						"time_zone":                  "US/Eastern",
						"locale_utc_offset":          float64(-240),
						"slack_username":             "somebody",
						"twitter_handle":             nil,
						"email":                      "somebody@withampersand.com",
						"email_client_email_address": "somebody@withampersand.com",
						"sending_email_address":      "somebody@withampersand.com",
						"from_address":               nil,
						"full_email_address":         "\"Int User\" <somebody@withampersand.com>",
						"bcc_email_address":          nil,
						"work_country":               nil,
						"seat_package":               "premier",
						"email_signature":            "",
						"email_signature_type":       "html",
						"email_signature_click_tracking_disabled": false,
						"team_admin":              true,
						"local_dial_enabled":      false,
						"click_to_call_enabled":   false,
						"email_client_configured": false,
						"crm_connected":           false,
						"external_feature_flags": map[string]any{
							"ma_enabled":               true,
							"ma_dev_qa_tools":          false,
							"ma_mobile_workflow":       true,
							"ma_dark_mode":             true,
							"hot_leads":                true,
							"people_crud_allow_create": true,
							"people_crud_allow_delete": true,
							"linkedin_oauth_flow":      true,
						},
						"_private_fields":         map[string]any{},
						"phone_client":            map[string]any{"id": float64(46885)},
						"phone_number_assignment": nil,
						"group":                   nil,
						"team": map[string]any{
							"_href": "https://api.salesloft.com/v2/team",
							"id":    float64(111855),
						},
						"role": map[string]any{
							"id": "Admin",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input)

			tt.Validate(t, err, result)
		})
	}
}
