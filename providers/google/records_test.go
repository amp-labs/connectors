package google

import (
	"net/http"
	"sort"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// TestCalendarGetRecordsByIds covers the Calendar subscribe fetch. recordIds[0] is the
// updatedMin checkpoint (not an event id); the method paginates events.list and only
// supports the "events" object.
func TestCalendarGetRecordsByIds(t *testing.T) { //nolint:funlen
	t.Parallel()

	eventsPage1 := testutils.DataFromFile(t, "calendar/records-by-ids/events-page-1.json")
	eventsPage2 := testutils.DataFromFile(t, "calendar/records-by-ids/events-page-2.json")

	const updatedMin = "2026-06-17T00:00:00.000Z"

	tests := []struct {
		name          string
		object        string
		ids           []string
		server        func() *mockserver.Switch
		expectErr     bool
		expectRowIds  []string // sorted
		expectNoError bool
	}{
		{
			name:   "Unsupported object returns ErrGetRecordNotSupportedForObject",
			object: "calendarList",
			ids:    []string{updatedMin},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{Setup: mockserver.ContentJSON()}
			},
			expectErr: true,
		},
		{
			name:   "Empty recordIds errors before hitting the provider",
			object: "events",
			ids:    []string{},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{Setup: mockserver.ContentJSON()}
			},
			expectErr: true,
		},
		{
			name:   "Blank updatedMin errors before hitting the provider",
			object: "events",
			ids:    []string{""},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{Setup: mockserver.ContentJSON()}
			},
			expectErr: true,
		},
		{
			name:   "Paginates across pages following nextPageToken",
			object: "events",
			ids:    []string{updatedMin},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Cases: []mockserver.Case{
						{
							// page-1 carries nextPageToken "TOKEN2"; the follow-up request
							// for page-2 is the only one carrying pageToken.
							If:   mockcond.QueryParam("pageToken", "TOKEN2"),
							Then: mockserver.Response(http.StatusOK, eventsPage2),
						},
					},
					Default: mockserver.Response(http.StatusOK, eventsPage1),
				}
			},
			expectRowIds:  []string{"evt-cancelled-1", "evt-confirmed-1"},
			expectNoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := tt.server().Server()
			t.Cleanup(srv.Close)

			conn, err := constructTestCalendarConnector(srv.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			rows, err := conn.GetRecordsByIds(t.Context(), tt.object, tt.ids, nil, nil)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected an error, got nil (rows: %v)", rows)
				}

				return
			}

			if tt.expectNoError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotIds := make([]string, 0, len(rows))
			for _, row := range rows {
				gotIds = append(gotIds, calendarRowID(row))
			}

			sort.Strings(gotIds)

			if len(gotIds) != len(tt.expectRowIds) {
				t.Fatalf("expected %d rows, got %d (ids: %v)", len(tt.expectRowIds), len(gotIds), gotIds)
			}

			for i, want := range tt.expectRowIds {
				if gotIds[i] != want {
					t.Fatalf("row %d: expected id %q, got %q (all: %v)", i, want, gotIds[i], gotIds)
				}
			}
		})
	}
}

// calendarRowID gets the event id from a row. Depending on which fields were requested it
// may sit on Id, in Fields, or only in Raw, so check all three.
func calendarRowID(row common.ReadResultRow) string {
	if row.Id != "" {
		return row.Id
	}

	if v, ok := row.Fields["id"].(string); ok && v != "" {
		return v
	}

	if v, ok := row.Raw["id"].(string); ok {
		return v
	}

	return ""
}
