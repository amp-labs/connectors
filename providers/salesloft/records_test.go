package salesloft

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// GetRecordsByIdsInput represents the input parameters for GetRecordsByIds method.
type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

func TestGetRecordByIds(t *testing.T) {
	t.Parallel()

	responseGetPeople := testutils.DataFromFile(t, "get-records-people.json")
	responseGetAccounts := testutils.DataFromFile(t, "get-records-accounts.json")
	responseGetUsers := testutils.DataFromFile(t, "read-list-users.json")

	tests := []testconn.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name:         "Missing object name returns error",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully fetch people by IDs",
			Input: GetRecordsByIdsInput{
				ObjectName: "people",
				Fields:     []string{"email_address"},
				Ids:        []string{"164510523", "164510464"},
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
			Input: GetRecordsByIdsInput{
				ObjectName: "accounts",
				Fields:     []string{"name"},
				Ids:        []string{"48371814", "48371806"},
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
			Input: GetRecordsByIdsInput{
				ObjectName: "users",
				Fields:     []string{"email"},
				Ids:        []string{"49067"},
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

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, nil)

			var rows []common.ReadResultRow
			if result != nil {
				rows = result.Rows
			}

			tt.Validate(t, err, rows)

			// Every requested id is in the response, so none may be reported missing.
			if err == nil {
				testconn.RequireNoFailures(t, result)
			}
		})
	}
}

// Ids the batch response leaves out are reported as not-found failures, in requested order.
func TestGetRecordsByIds_ReportsMissingIds(t *testing.T) {
	t.Parallel()

	response := testutils.DataFromFile(t, "get-records-people.json")

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/v2/people"),
		Then:  mockserver.Response(http.StatusOK, response),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// The response holds "164510523", "164510464"; the other ids are absent from it.
	result, err := conn.GetRecordsByIds(t.Context(), "people",
		[]string{"164510001", "164510523", "164510002", "164510464"}, []string{"id"}, nil)
	if err != nil {
		t.Fatalf("a partial batch response must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "164510523", "164510464")
	testconn.RequireMissingIds(t, result, "164510001", "164510002")
}

// makeIds returns count ids starting at first, as strings.
func makeIds(first, count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = strconv.Itoa(first + i)
	}

	return ids
}

// More ids than fit on one page are requested in groups of DefaultPageSize, and the rows of every
// group are returned.
func TestGetRecordsByIds_SplitsIdsIntoPages(t *testing.T) {
	t.Parallel()

	ids := makeIds(1001, 150)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.And{
				mockcond.Path("/v2/people"),
				mockcond.QueryParam("per_page", "100"),
				mockcond.QueryParam("ids[]", ids[:100]...),
			},
			Then: mockserver.ResponseString(http.StatusOK, `{"data":[{"id":1001},{"id":1100}]}`),
		}, {
			If: mockcond.And{
				mockcond.Path("/v2/people"),
				mockcond.QueryParam("per_page", "100"),
				mockcond.QueryParam("ids[]", ids[100:]...),
			},
			Then: mockserver.ResponseString(http.StatusOK, `{"data":[{"id":1101},{"id":1150}]}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "people", ids, []string{"id"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testconn.RequireRowIds(t, result, "1001", "1100", "1101", "1150")

	// Every id that neither response returned is reported as not found.
	var missing []string

	for _, id := range ids {
		if !slices.Contains([]string{"1001", "1100", "1101", "1150"}, id) {
			missing = append(missing, id)
		}
	}

	testconn.RequireMissingIds(t, result, missing...)
}

// When one group of ids fails, its ids are reported as failures and returned in the error, while the
// rows of the other group are kept.
func TestGetRecordsByIds_FailedPageKeepsOtherRows(t *testing.T) {
	t.Parallel()

	ids := makeIds(1001, 150)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.And{
				mockcond.Path("/v2/people"),
				mockcond.QueryParam("ids[]", ids[:100]...),
			},
			Then: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"Internal Server Error"}`),
		}, {
			If: mockcond.And{
				mockcond.Path("/v2/people"),
				mockcond.QueryParam("ids[]", ids[100:]...),
			},
			Then: mockserver.ResponseString(http.StatusOK, `{"data":[{"id":1101},{"id":1150}]}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "people", ids, []string{"id"}, nil)
	if !errors.Is(err, common.ErrServer) {
		t.Fatalf("expected the failed request's error, got %v", err)
	}

	testconn.RequireRowIds(t, result, "1101", "1150")

	failedIds := make([]string, 0, len(result.Failures))
	notFoundIds := make([]string, 0, len(result.Failures))

	for _, failure := range result.Failures {
		if failure.Reason == common.FailureReasonNotFound {
			notFoundIds = append(notFoundIds, failure.RecordId)
		} else {
			failedIds = append(failedIds, failure.RecordId)
		}
	}

	if !slices.Equal(failedIds, ids[:100]) {
		t.Errorf("failed ids %v, want the first 100 requested ids", failedIds)
	}

	// The second group returned 1101 and 1150; its other ids are not found.
	wantNotFound := slices.DeleteFunc(slices.Clone(ids[100:]), func(id string) bool {
		return id == "1101" || id == "1150"
	})
	if !slices.Equal(notFoundIds, wantNotFound) {
		t.Errorf("not-found ids %v, want %v", notFoundIds, wantNotFound)
	}
}

// A 404 still fails the whole call with the provider's error, as it did before ids were split into
// groups.
func TestGetRecordsByIds_NotFoundFailsCall(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/v2/people"),
		Then:  mockserver.ResponseString(http.StatusNotFound, `{"error":"Not Found"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "people", []string{"1001", "1002"}, []string{"id"}, nil)
	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("expected a not-found error, got %v", err)
	}

	if result != nil {
		t.Errorf("expected no result, got %+v", result)
	}
}
