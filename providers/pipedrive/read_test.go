package pipedrive

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	nextPageTest := testutils.DataFromFile(t, "activities.json")
	leads := testutils.DataFromFile(t, "leads.json")
	ErrResponseBody := testutils.DataFromFile(t, "not-found.json")

	server := mockserver.Fixed{
		Setup:  mockserver.ContentJSON(),
		Always: mockserver.Response(http.StatusOK, nextPageTest),
	}.Server()

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
			Comparator:   testroutines.ComparatorSubsetRead,
			ExpectedErrs: nil,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": float64(2),
						},
						Raw: map[string]any{
							"id":                  float64(2),
							"company_id":          float64(13313052),
							"user_id":             float64(20580207),
							"done":                false,
							"type":                "call",
							"due_date":            "2024-10-30",
							"busy_flag":           false,
							"add_time":            "2024-10-16 12:16:02",
							"subject":             "I usually can't come up with words",
							"public_description":  "Demo activity",
							"location":            "Dar es salaam",
							"active_flag":         true,
							"update_time":         "2024-10-16 12:16:02",
							"private":             false,
							"created_by_user_id":  float64(20580207),
							"owner_name":          "Integration User",
							"assigned_to_user_id": float64(20580207),
							"type_name":           "Call",
						},
					},
				},
				Done:     false,
				NextPage: common.NextPageToken(server.URL + "/v1/activities?start=1"),
			},
		},
		{
			Name:  "Resource not found",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, ErrResponseBody),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				testutils.StringError(string(ErrResponseBody)),
			},
			Expected: nil,
		},
		{
			Name: "Successful read of leads",
			Input: common.ReadParams{
				ObjectName: "leads",
				Fields:     connectors.Fields("channel", "id", "origin", "title"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/leads"),
				Then:  mockserver.Response(http.StatusOK, leads),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"channel": nil,
						"id":      "8ba414d0-8ad8-11ef-9e9e-07879dd74146",
						"origin":  "ManuallyCreated",
						"title":   "Ampersand lead",
					},
					Raw: map[string]any{
						"add_time":        "2024-10-15T09:33:33.213Z",
						"cc_email":        "integrationuser-sandbox2+13313052+leadif7ZKQ7mhWz28LivD9BmE3@pipedrivemail.com",
						"creator_id":      float64(20580207),
						"id":              "8ba414d0-8ad8-11ef-9e9e-07879dd74146",
						"is_archived":     false,
						"label_ids":       []any{},
						"organization_id": float64(2),
						"origin":          "ManuallyCreated",
						"channel":         nil,
						"owner_id":        float64(20580207),
						"person_id":       float64(2),
						"source_name":     "Manually created",
						"title":           "Ampersand lead",
						"update_time":     "2024-10-15T09:33:33.213Z",
						"visible_to":      "3",
						"was_seen":        true,
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL, providers.ModulePipedriveLegacy)
			})
		})
	}
}
