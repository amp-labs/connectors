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
	responseNote := testutils.DataFromFile(t, "note.json")
	tests := []testroutines.Read{
		{
			Name: "Read empty items",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title", "created_at"),
				PageSize:   4,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/notes"),
					mockcond.QueryParam("page_size", "4"),
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
				PageSize:   4,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/v1/notes"),
							mockcond.QueryParam("page_size", "4"),
						},
						Then: mockserver.Response(http.StatusOK, responseNotes),
					},
					{
						If:   mockcond.Path("/v1/notes/not_1d3tmYTlCICgjy"),
						Then: mockserver.Response(http.StatusOK, responseNote),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
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
				Fields:     connectors.Fields("id", "title"),
				NextPage:   "eyJjcmVkZW50aWFsfQ==",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/v1/notes"),
							mockcond.QueryParam("cursor", "eyJjcmVkZW50aWFsfQ=="),
						},
						Then: mockserver.Response(http.StatusOK, responseNotesLastPage),
					},
					{
						If:   mockcond.Path("/v1/notes/not_1d3tmYTlCICgjy"),
						Then: mockserver.Response(http.StatusOK, responseNote),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "not_1d3tmYTlCICgjy",
							"title": "Quarterly yoghurt budget review",
						},
						Raw: map[string]any{
							"id":    "not_1d3tmYTlCICgjy",
							"title": "Quarterly yoghurt budget review",
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
				PageSize:   4,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/v1/notes"),
							mockcond.QueryParam("page_size", "4"),
						},
						Then: mockserver.Response(http.StatusOK, responseNotes),
					},
					{
						If:   mockcond.Path("/v1/notes/not_1d3tmYTlCICgjy"),
						Then: mockserver.Response(http.StatusOK, responseNote),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "not_1d3tmYTlCICgjy", "title": "Quarterly yoghurt budget review"},
						Raw:    map[string]any{"id": "not_1d3tmYTlCICgjy", "object": "note", "title": "Quarterly yoghurt budget review", "owner": map[string]any{"name": "Oat Benson", "email": "oat@granola.ai"}, "created_at": "2026-01-27T15:30:00Z"},
					},
				},
				NextPage: "eyJjcmVkZW50aWFsfQ==",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notes with Since and Until adds updated_after and updated_before query params",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("id", "title"),
				PageSize:   4,
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/v1/notes"),
							mockcond.QueryParam("page_size", "4"),
							mockcond.QueryParam("updated_after", "2026-01-01T00:00:00Z"),
							mockcond.QueryParam("updated_before", "2026-01-31T23:59:59Z"),
						},
						Then: mockserver.Response(http.StatusOK, responseNotes),
					},
					{
						If:   mockcond.Path("/v1/notes/not_1d3tmYTlCICgjy"),
						Then: mockserver.Response(http.StatusOK, responseNote),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "not_1d3tmYTlCICgjy", "title": "Quarterly yoghurt budget review"},
						Raw:    map[string]any{"id": "not_1d3tmYTlCICgjy", "object": "note", "title": "Quarterly yoghurt budget review", "owner": map[string]any{"name": "Oat Benson", "email": "oat@granola.ai"}, "created_at": "2026-01-27T15:30:00Z"},
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
