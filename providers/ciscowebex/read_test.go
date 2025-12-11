package ciscowebex

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

// nolint:funlen,gocognit,cyclop
func TestRead(t *testing.T) {
	t.Parallel()

	responseInvalidPath := testutils.DataFromFile(t, "invalid-path.json")
	responseReadEmpty := testutils.DataFromFile(t, "read-empty.json")
	responseReadPeople := testutils.DataFromFile(t, "read-people.json")
	responseReadPeopleFirstPage := testutils.DataFromFile(t, "read-people-first-page.json")
	responseReadPeopleSecondPage := testutils.DataFromFile(t, "read-people-second-page.json")
	responseReadGroupsFiltered := testutils.DataFromFile(t, "read-groups-filtered.json")
	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "people"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unknown objects return HTTP error",
			Input: common.ReadParams{ObjectName: "unknown", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/unknown"),
				Then:  mockserver.Response(http.StatusTeapot),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
		{
			Name: "Read invalid path",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName"),
				NextPage:   testroutines.URLTestServer + "/v1/invalidpath",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/invalidpath"),
				Then:  mockserver.Response(http.StatusNotFound, responseInvalidPath),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
			},
		},
		{
			Name: "Read empty items",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseReadEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read people",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName", "emails", "firstName", "lastName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/people"),
				Then:  mockserver.Response(http.StatusOK, responseReadPeople),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayname": "testuser",
							"emails":      []any{"testuser@example.com"},
							"firstname":   "test",
							"lastname":    "user",
						},
						Raw: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayName": "testuser",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read people first page with pagination",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName", "emails"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/people"),
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link", `<https://webexapis.com/v1/people?max=1&cursor=next_cursor_token>; rel="next"`),
					mockserver.Response(http.StatusOK, responseReadPeopleFirstPage),
				),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL1BFT",
							"displayname": "admin@example.wbx.ai",
							"emails":      []any{"admin@example.wbx.ai"},
						},
						Raw: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL1BFT",
							"displayName": "admin@example.wbx.ai",
							"emails":      []any{"admin@example.wbx.ai"},
							"nickName":    "admin",
							"firstName":   "admin",
							"lastName":    "admin",
							"orgId":       "Y2lzY29zcGFyazovL3VzL09",
							"type":        "person",
						},
					},
				},
				NextPage: "https://webexapis.com/v1/people?max=1&cursor=next_cursor_token",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read people second page using NextPage token",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName", "emails"),
				NextPage:   testroutines.URLTestServer + "/v1/people?max=1&cursor=next_cursor_token",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/people"),
					mockcond.QueryParam("cursor", "next_cursor_token"),
				},
				Then: mockserver.Response(http.StatusOK, responseReadPeopleSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayname": "testuser",
							"emails":      []any{"testuser@example.com"},
						},
						Raw: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayName": "testuser",
							"emails":      []any{"testuser@example.com"},
							"nickName":    "testuser",
							"firstName":   "test",
							"lastName":    "user",
							"orgId":       "Y2lzY29zcGFyazovL3",
							"type":        "person",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read people with PageSize uses max query param",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/people"),
					mockcond.QueryParam("max", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseReadPeople),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayname": "testuser",
						},
						Raw: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayName": "testuser",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read groups with PageSize uses count query param",
			Input: common.ReadParams{
				ObjectName: "groups",
				Fields:     connectors.Fields("id", "displayName"),
				PageSize:   25,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/groups"),
					mockcond.QueryParam("count", "25"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read-groups.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcyZj",
							"displayname": "Site1",
						},
						Raw: map[string]any{
							"id":           "Y2lzY29zcyZj",
							"displayName":  "Site1",
							"orgId":        "Y2lzY29zcGF",
							"created":      "2025-12-05T21:16:52.911Z",
							"lastModified": "2025-12-05T21:33:48.280Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read groups with Since filters records connector-side",
			Input: common.ReadParams{
				ObjectName: "groups",
				Fields:     connectors.Fields("id", "displayName"),
				Since:      time.Date(2025, 12, 5, 21, 30, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/groups"),
				Then:  mockserver.Response(http.StatusOK, responseReadGroupsFiltered),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcyZj",
							"displayname": "Site1",
						},
						Raw: map[string]any{
							"id":           "Y2lzY29zcyZj",
							"displayName":  "Site1",
							"orgId":        "Y2lzY29zcGF",
							"created":      "2025-12-05T21:16:52.911Z",
							"lastModified": "2025-12-05T21:33:48.280Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read groups with Since before all records returns empty",
			Input: common.ReadParams{
				ObjectName: "groups",
				Fields:     connectors.Fields("id", "displayName"),
				Since:      time.Date(2025, 12, 5, 22, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/groups"),
				Then:  mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read-groups.json")),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read people with Since does not filter (no time-based filtering)",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("id", "displayName"),
				Since:      time.Date(2025, 12, 5, 22, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/people"),
				Then:  mockserver.Response(http.StatusOK, responseReadPeople),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayname": "testuser",
						},
						Raw: map[string]any{
							"id":          "Y2lzY29zcGFyazovL3VzL",
							"displayName": "testuser",
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
