package breakcold

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	leadResponse := testutils.DataFromFile(t, "write_lead.json")
	remindersResponse := testutils.DataFromFile(t, "write_reminders.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the lead",
			Input: common.WriteParams{ObjectName: "lead", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/lead"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, leadResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "fe3b2787-dc05-4f45-942d-a45ee960d9c8",
				Errors:   nil,
				Data: map[string]any{
					"id":                "fe3b2787-dc05-4f45-942d-a45ee960d9c8",
					"email":             "google",
					"company":           "google",
					"search_query":      "google google google google califonia provider ",
					"created_at":        "2025-08-14T08:26:39.881Z",
					"updated_at":        nil,
					"first_name":        "google",
					"last_name":         "google",
					"city":              "Califonia",
					"company_role":      "provider",
					"contract_currency": "USD",
					"id_space":          "a5bf4d9d-46d3-42f4-b759-7bace001ea1b",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update lead as PATCH",
			Input: common.WriteParams{
				ObjectName: "lead",
				RecordId:   "fe3b2787-dc05-4f45-942d-a45ee960d9c8",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/leads/fe3b2787-dc05-4f45-942d-a45ee960d9c8"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, leadResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "fe3b2787-dc05-4f45-942d-a45ee960d9c8",
				Errors:   nil,
				Data: map[string]any{
					"id":                "fe3b2787-dc05-4f45-942d-a45ee960d9c8",
					"email":             "google",
					"company":           "google",
					"search_query":      "google google google google califonia provider ",
					"created_at":        "2025-08-14T08:26:39.881Z",
					"updated_at":        nil,
					"first_name":        "google",
					"last_name":         "google",
					"city":              "Califonia",
					"company_role":      "provider",
					"contract_currency": "USD",
					"id_space":          "a5bf4d9d-46d3-42f4-b759-7bace001ea1b",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the reminders",
			Input: common.WriteParams{ObjectName: "reminders", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/reminders"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, remindersResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "be392c6c-d785-435b-9eb8-d4988f025160",
				Errors:   nil,
				Data: map[string]any{
					"id":         "be392c6c-d785-435b-9eb8-d4988f025160",
					"date":       nil,
					"name":       "reminder call",
					"cron":       nil,
					"cron_id":    nil,
					"created_at": "2025-08-14T09:41:36.969Z",
					"users": []any{
						map[string]any{
							"avatar_url": nil,
							"handle":     "sample257",
							"email":      "sample@gmail.com",
							"id":         "8XZHWCI3EVaPfiZpYL5shZRAwjj2",
							"full_name":  "Sample",
							"notify":     false,
							"is_author":  true,
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update reminders as PATCH",
			Input: common.WriteParams{
				ObjectName: "reminders",
				RecordId:   "be392c6c-d785-435b-9eb8-d4988f025160",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/reminders/be392c6c-d785-435b-9eb8-d4988f025160"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, remindersResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "be392c6c-d785-435b-9eb8-d4988f025160",
				Errors:   nil,
				Data: map[string]any{
					"id":         "be392c6c-d785-435b-9eb8-d4988f025160",
					"date":       nil,
					"name":       "reminder call",
					"cron":       nil,
					"cron_id":    nil,
					"created_at": "2025-08-14T09:41:36.969Z",
					"users": []any{
						map[string]any{
							"avatar_url": nil,
							"handle":     "sample257",
							"email":      "sample@gmail.com",
							"id":         "8XZHWCI3EVaPfiZpYL5shZRAwjj2",
							"full_name":  "Sample",
							"notify":     false,
							"is_author":  true,
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

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
