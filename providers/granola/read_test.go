package granola

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

func TestRead(t *testing.T) {
	t.Parallel()
	responseReadEmpty := testutils.DataFromFile(t, "read-empty.json")
	responseNotes := testutils.DataFromFile(t, "notes.json")
	responseNotesLastPage := testutils.DataFromFile(t, "notes-last-page.json")

	tests := []testroutines.Read{
		{
			Name: "Read empty items",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title", "created_at"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseReadEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notes",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title", "created_at"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "10"),
				},
				Then: mockserver.Response(http.StatusOK, responseNotes),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":         "not_1d3tmYTlCICgjy",
							"title":      "Quarterly yoghurt budget review",
							"created_at": "2026-01-27T15:30:00Z",
						},
						Raw: map[string]any{
							"id":         "not_1d3tmYTlCICgjy",
							"object":     "note",
							"title":      "Quarterly yoghurt budget review",
							"owner":      map[string]any{"name": "Oat Benson", "email": "oat@granola.ai"},
							"created_at": "2026-01-27T15:30:00Z",
						},
					},
					{
						Fields: map[string]any{
							"id":         "not_7hKpQ2LmZx91ab",
							"title":      "Customer onboarding flow improvements",
							"created_at": "2026-01-29T10:15:00Z",
						},
						Raw: map[string]any{
							"id":         "not_7hKpQ2LmZx91ab",
							"object":     "note",
							"title":      "Customer onboarding flow improvements",
							"owner":      map[string]any{"name": "Maya Tesfaye", "email": "maya.tesfaye@granola.ai"},
							"created_at": "2026-01-29T10:15:00Z",
						},
					},
					{
						Fields: map[string]any{
							"id":         "not_4Js8PdLwQe72rt",
							"title":      "transaction latency analysis",
							"created_at": "2026-02-02T08:45:00Z",
						},
						Raw: map[string]any{
							"id":         "not_4Js8PdLwQe72rt",
							"object":     "note",
							"title":      "transaction latency analysis",
							"owner":      map[string]any{"name": "Eren Yeager", "email": "eren.yeager@granola.ai"},
							"created_at": "2026-02-02T08:45:00Z",
						},
					},
				},
				NextPage: "eyJjcmVkZW50aWFsfQ==",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notes second page using NextPage token",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title", "created_at"),
				NextPage:   "eyJjcmVkZW50aWFsfQ==",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("cursor", "eyJjcmVkZW50aWFsfQ=="),
				},
				Then: mockserver.Response(http.StatusOK, responseNotesLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":         "not_8xYzAbcDef",
							"title":      "Final note",
							"created_at": "2026-02-10T12:00:00Z",
						},
						Raw: map[string]any{
							"id":         "not_8xYzAbcDef",
							"object":     "note",
							"title":      "Final note",
							"owner":      map[string]any{"name": "Test User", "email": "test@granola.ai"},
							"created_at": "2026-02-10T12:00:00Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notes with PageSize uses page_size query param",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseNotes),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "not_1d3tmYTlCICgjy", "title": "Quarterly yoghurt budget review"},
						Raw:    map[string]any{"id": "not_1d3tmYTlCICgjy", "object": "note", "title": "Quarterly yoghurt budget review", "owner": map[string]any{"name": "Oat Benson", "email": "oat@granola.ai"}, "created_at": "2026-01-27T15:30:00Z"},
					},
					{
						Fields: map[string]any{"id": "not_7hKpQ2LmZx91ab", "title": "Customer onboarding flow improvements"},
						Raw:    map[string]any{"id": "not_7hKpQ2LmZx91ab", "object": "note", "title": "Customer onboarding flow improvements", "owner": map[string]any{"name": "Maya Tesfaye", "email": "maya.tesfaye@granola.ai"}, "created_at": "2026-01-29T10:15:00Z"},
					},
					{
						Fields: map[string]any{"id": "not_4Js8PdLwQe72rt", "title": "transaction latency analysis"},
						Raw:    map[string]any{"id": "not_4Js8PdLwQe72rt", "object": "note", "title": "transaction latency analysis", "owner": map[string]any{"name": "Eren Yeager", "email": "eren.yeager@granola.ai"}, "created_at": "2026-02-02T08:45:00Z"},
					},
				},
				NextPage: "eyJjcmVkZW50aWFsfQ==",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notes with Since and Until adds created_after and created_before query params",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "10"),
					mockcond.QueryParam("created_after", "2026-01-01T00:00:00Z"),
					mockcond.QueryParam("created_before", "2026-01-31T23:59:59Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseNotes),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "not_1d3tmYTlCICgjy", "title": "Quarterly yoghurt budget review"},
						Raw:    map[string]any{"id": "not_1d3tmYTlCICgjy", "object": "note", "title": "Quarterly yoghurt budget review", "owner": map[string]any{"name": "Oat Benson", "email": "oat@granola.ai"}, "created_at": "2026-01-27T15:30:00Z"},
					},
					{
						Fields: map[string]any{"id": "not_7hKpQ2LmZx91ab", "title": "Customer onboarding flow improvements"},
						Raw:    map[string]any{"id": "not_7hKpQ2LmZx91ab", "object": "note", "title": "Customer onboarding flow improvements", "owner": map[string]any{"name": "Maya Tesfaye", "email": "maya.tesfaye@granola.ai"}, "created_at": "2026-01-29T10:15:00Z"},
					},
					{
						Fields: map[string]any{"id": "not_4Js8PdLwQe72rt", "title": "transaction latency analysis"},
						Raw:    map[string]any{"id": "not_4Js8PdLwQe72rt", "object": "note", "title": "transaction latency analysis", "owner": map[string]any{"name": "Eren Yeager", "email": "eren.yeager@granola.ai"}, "created_at": "2026-02-02T08:45:00Z"},
					},
				},
				NextPage: "eyJjcmVkZW50aWFsfQ==",
				Done:     false,
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
