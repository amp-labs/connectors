package lever

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	opportunitiesResponse := testutils.DataFromFile(t, "opportunities.json")
	requisitionFieldsResponse := testutils.DataFromFile(t, "requisition_fields.json")
	usersResponse := testutils.DataFromFile(t, "users.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read list of opportunities",
			Input: common.ReadParams{
				ObjectName: "opportunities",
				Fields:     connectors.Fields(""),
				Since:      time.Date(2025, time.June, 2, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2025, time.June, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/opportunities"),
				Then:  mockserver.Response(http.StatusOK, opportunitiesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":              "2087af84-f146-4535-9368-2309e33e049f",
							"name":            "karan",
							"contact":         "e2719ee1-21f5-4287-a60a-59bd29deac80",
							"headline":        "",
							"stage":           "lead-new",
							"confidentiality": "non-confidential",
							"location":        "",
							"phones":          []any{},
							"emails":          []any{},
							"links":           []any{},
							"archived":        nil,
							"tags": []any{
								"Hiring Software developer",
								"Go tech",
								"kovilpatti",
								"Service based",
								"hiring",
							},
							"sources": []any{
								"Added manually",
							},
							"stageChanges": []any{
								map[string]any{
									"toStageId":    "lead-new",
									"toStageIndex": float64(0),
									"updatedAt":    float64(1750233426190),
									"userId":       "2c713392-d0e4-4355-8673-378d6a851cb8",
								},
							},
							"origin":    "sourced",
							"sourcedBy": "2c713392-d0e4-4355-8673-378d6a851cb8",
							"owner":     "2c713392-d0e4-4355-8673-378d6a851cb8",
							"followers": []any{
								"2c713392-d0e4-4355-8673-378d6a851cb8",
								"e0a9f3e3-bd9b-4eb1-8d2c-0d9a572bf641",
							},
							"applications": []any{
								"2b7f3433-111a-4659-a302-58a32cd2e33c",
							},
							"createdAt":         float64(1750233426190),
							"updatedAt":         float64(1750247948243),
							"lastInteractionAt": float64(1750247948074),
							"lastAdvancedAt":    float64(1750233426190),
							"snoozedUntil":      nil,
							"urls": map[string]any{
								"list": "https://hire.sandbox.lever.co/candidates",
								"show": "https://hire.sandbox.lever.co/candidates/2087af84-f146-4535-9368-2309e33e049f",
							},
							"isAnonymized":        false,
							"dataProtection":      nil,
							"opportunityLocation": "Chennai",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v1/opportunities?limit=100" +
					"&updated_at_start=1748822400000&updated_at_end=1751241600000" +
					"&offset=%255B1%252C1750233409740%252C%252248dd4e94-fea0-4f9a-be5f-95b1853cbbbe%2522%255D",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of Requisition fields",
			Input: common.ReadParams{ObjectName: "requisition_fields", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/requisition_fields"),
				Then:  mockserver.Response(http.StatusOK, requisitionFieldsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": "field1",
						},
						Raw: map[string]any{
							"id":         "field1",
							"text":       "Area of Interest",
							"type":       "text",
							"isRequired": false,
						},
						Id: "field1",
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of users",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/users"),
				Then:  mockserver.Response(http.StatusOK, usersResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":                  "e0a9f3e3-bd9b-4eb1-8d2c-0d9a572bf641",
							"name":                "sample@gmail.com",
							"username":            "sample",
							"email":               "sample@gmail.com",
							"accessRole":          "super admin",
							"photo":               nil,
							"createdAt":           float64(1743625450829),
							"deactivatedAt":       nil,
							"externalDirectoryId": nil,
							"linkedContactIds":    nil,
							"jobTitle":            nil,
							"managerId":           nil,
						},
					},
				},
				NextPage: testroutines.URLTestServer +
					"/v1/users?limit=100&offset=%255B1733206956064%252C%2522c21d911f-8292-49b7-b135-f7cc233b43fd%2522%255D",
				Done: false,
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
