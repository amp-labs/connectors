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

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	statusResponse := testutils.DataFromFile(t, "status.json")
	remindersResponse := testutils.DataFromFile(t, "reminders.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of status",
			Input: common.ReadParams{ObjectName: "status", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/status"),
				Then:  mockserver.Response(http.StatusOK, statusResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"type":         nil,
							"id":           "24a13522-d6fd-48fc-be9b-e9b331e5f194",
							"name":         "Engaged",
							"order":        float64(0),
							"color":        "#8ed1fc",
							"success_rate": float64(10),
							"id_space":     "a5bf4d9d-46d3-42f4-b759-7bace001ea1b",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of reminders list",
			Input: common.ReadParams{ObjectName: "reminders/list", Fields: connectors.Fields(""), NextPage: "2"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/reminders/list"),
				Then:  mockserver.Response(http.StatusOK, remindersResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":      "7312bd6f-a5c7-431d-83f4-13fa9c7257c6",
							"date":    nil,
							"name":    "demo20",
							"cron":    nil,
							"cron_id": nil,
							"users": []any{
								map[string]any{
									"id_lead_reminder": "7312bd6f-a5c7-431d-83f4-13fa9c7257c6",
									"id_user":          "8XZHWCI3EVaPfiZpYL5shZRAwjj2",
									"assigned_at":      "2025-08-13T10:44:20.832Z",
									"is_author":        true,
									"notify":           false,
									"user": map[string]any{
										"id":        "8XZHWCI3EVaPfiZpYL5shZRAwjj2",
										"email":     "sample@gmail.com",
										"full_name": "sample",
									},
								},
							},
						},
					},
				},
				Done: true,
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
