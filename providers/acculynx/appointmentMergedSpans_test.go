package acculynx

import (
	"testing"
	"time"

	"github.com/amp-labs/amp-common/feature"
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"gotest.tools/v3/assert"
)

// TestReadMergedAppointmentSpans covers calendars/appointments reads with
// appointmentMergedSpansFlag on, with appointmentOccurrencesFlag off and on. It
// is not parallel because both flags are process-global: TestRead's parallel
// cases, which read them off, resume only after this test has returned and
// reset them.
//
//nolint:paralleltest,tparallel,funlen // flips the process-global appointment flags
func TestReadMergedAppointmentSpans(t *testing.T) {
	assert.NilError(t, appointmentMergedSpansFlag.Set(feature.On()))
	t.Cleanup(appointmentMergedSpansFlag.Reset)
	t.Cleanup(appointmentOccurrencesFlag.Reset)

	tests := []struct {
		occurrences bool
		testconn.TestCaseRead
	}{
		{
			// Occurrence flag off: the rows and their ids are AccuLynx's as they
			// come, and each row carries its appointment's span, unrequested.
			TestCaseRead: testconn.TestCaseRead{
				Name: "Merged-span read of calendars/appointments adds the appointment's span to each row",
				Input: common.ReadParams{
					ObjectName: "calendars/appointments",
					Fields:     connectors.Fields("id"),
					Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
					Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
				},
				Server: appointmentsServer(appointmentsWindowCase("2026-04-01", "2026-04-30",
					appointmentsPage(
						`{"id": "crew-1", "start": "2026-04-01T13:00:00Z", "end": "2026-04-01T22:00:00Z"}`,
						`{"id": "crew-1", "start": "2026-04-02T13:00:00Z", "end": "2026-04-02T22:00:00Z"}`,
						`{"id": "appt-1", "start": "2026-04-15T14:00:00Z", "end": "2026-04-15T15:00:00Z"}`,
					))),
				Comparator: testconn.ComparatorSubsetRead,
				Expected: &common.ReadResult{
					Rows: 3,
					Data: []common.ReadResultRow{
						{
							Id: "crew-1",
							Fields: map[string]any{
								"id":               "crew-1",
								"appointmentstart": "2026-04-01T13:00:00Z",
								"appointmentend":   "2026-04-02T22:00:00Z",
							},
							Raw: map[string]any{"id": "crew-1", "start": "2026-04-01T13:00:00Z"},
						},
						{
							Id: "crew-1",
							Fields: map[string]any{
								"id":               "crew-1",
								"appointmentstart": "2026-04-01T13:00:00Z",
								"appointmentend":   "2026-04-02T22:00:00Z",
							},
							Raw: map[string]any{"id": "crew-1", "start": "2026-04-02T13:00:00Z"},
						},
						{
							Id: "appt-1",
							Fields: map[string]any{
								"id":               "appt-1",
								"appointmentstart": "2026-04-15T14:00:00Z",
								"appointmentend":   "2026-04-15T15:00:00Z",
							},
							Raw: map[string]any{"id": "appt-1"},
						},
					},
					Done: true,
				},
			},
		},
		{
			// Occurrence flag on: each day keeps its occurrence id, the repeat
			// returned by both windows is merged, and the span covers the days
			// of both windows.
			occurrences: true,
			TestCaseRead: testconn.TestCaseRead{
				Name: "Merged-span read of calendars/appointments with occurrences spans the occurrences of every window",
				Input: common.ReadParams{
					ObjectName: "calendars/appointments",
					Fields:     connectors.Fields("id"),
					Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
				},
				Server: appointmentsServer(
					appointmentsWindowCase("2026-01-01", "2026-03-31",
						appointmentsPage(
							`{"id": "crew-1", "start": "2026-03-31T13:00:00Z", "end": "2026-03-31T22:00:00Z"}`,
						)),
					appointmentsWindowCase("2026-04-01", "2026-04-30",
						appointmentsPage(
							`{"id": "crew-1", "start": "2026-03-31T13:00:00Z", "end": "2026-03-31T22:00:00Z"}`,
							`{"id": "crew-1", "start": "2026-04-01T13:00:00Z", "end": "2026-04-01T22:00:00Z"}`,
						)),
				),
				Comparator: testconn.ComparatorSubsetRead,
				Expected: &common.ReadResult{
					Rows: 2,
					Data: []common.ReadResultRow{
						{
							Id: "crew-1_2026-03-31T13:00:00Z",
							Fields: map[string]any{
								"id":               "crew-1_2026-03-31T13:00:00Z",
								"appointmentstart": "2026-03-31T13:00:00Z",
								"appointmentend":   "2026-04-01T22:00:00Z",
							},
							Raw: map[string]any{"id": "crew-1", "start": "2026-03-31T13:00:00Z"},
						},
						{
							Id: "crew-1_2026-04-01T13:00:00Z",
							Fields: map[string]any{
								"id":               "crew-1_2026-04-01T13:00:00Z",
								"appointmentstart": "2026-03-31T13:00:00Z",
								"appointmentend":   "2026-04-01T22:00:00Z",
							},
							Raw: map[string]any{"id": "crew-1", "start": "2026-04-01T13:00:00Z"},
						},
					},
					Done: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.occurrences {
				assert.NilError(t, appointmentOccurrencesFlag.Set(feature.On()))
			} else {
				appointmentOccurrencesFlag.Reset()
			}

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestReadConnector(tt.Server.URL)
			})
		})
	}
}

// TestReadsMergedAppointmentSpans checks that only calendars/appointments reads
// follow appointmentMergedSpansFlag.
//
//nolint:paralleltest // flips the process-global appointmentMergedSpansFlag
func TestReadsMergedAppointmentSpans(t *testing.T) {
	t.Cleanup(appointmentMergedSpansFlag.Reset)

	appointments := common.ReadParams{ObjectName: "calendars/appointments"}

	assert.Assert(t, !readsMergedAppointmentSpans(t.Context(), appointments), "off by default")

	assert.NilError(t, appointmentMergedSpansFlag.Set(feature.On()))

	assert.Assert(t, readsMergedAppointmentSpans(t.Context(), appointments), "on for calendars/appointments")
	assert.Assert(t, !readsMergedAppointmentSpans(t.Context(), common.ReadParams{ObjectName: "calendars"}),
		"other objects never add spans")
}

// TestAddMergedAppointmentSpans pins the merged span of each appointment: times
// compared as instants but kept as AccuLynx wrote them, a field left out when
// no row has a usable value for it, rows without an id untouched, and the
// rows' order, ids and Raw unchanged.
func TestAddMergedAppointmentSpans(t *testing.T) { //nolint:funlen
	t.Parallel()

	rows := []common.ReadResultRow{
		{
			Id:     "a",
			Fields: map[string]any{"id": "a"},
			Raw:    map[string]any{"id": "a", "start": "2026-04-01T09:00:00-04:00", "end": "2026-04-01T17:00:00-04:00"},
		},
		{
			Id:     "no-id",
			Fields: map[string]any{},
			Raw:    map[string]any{"start": "2026-04-01T09:00:00Z"},
		},
		{
			// 12:00Z is earlier than 09:00-04:00 (13:00Z), though it sorts
			// later as text.
			Id:     "a",
			Fields: map[string]any{"id": "a"},
			Raw:    map[string]any{"id": "a", "start": "2026-04-01T12:00:00Z", "end": "2026-04-02T21:30:00Z"},
		},
		{
			Id:  "b",
			Raw: map[string]any{"id": "b", "start": "2026-04-03T08:00:00Z", "end": "soon"},
		},
	}

	addMergedAppointmentSpans(rows)

	assert.DeepEqual(t, rows, []common.ReadResultRow{
		{
			Id: "a",
			Fields: map[string]any{
				"id":               "a",
				"appointmentstart": "2026-04-01T12:00:00Z",
				"appointmentend":   "2026-04-02T21:30:00Z",
			},
			Raw: map[string]any{"id": "a", "start": "2026-04-01T09:00:00-04:00", "end": "2026-04-01T17:00:00-04:00"},
		},
		{
			Id:     "no-id",
			Fields: map[string]any{},
			Raw:    map[string]any{"start": "2026-04-01T09:00:00Z"},
		},
		{
			Id: "a",
			Fields: map[string]any{
				"id":               "a",
				"appointmentstart": "2026-04-01T12:00:00Z",
				"appointmentend":   "2026-04-02T21:30:00Z",
			},
			Raw: map[string]any{"id": "a", "start": "2026-04-01T12:00:00Z", "end": "2026-04-02T21:30:00Z"},
		},
		{
			Id:     "b",
			Fields: map[string]any{"appointmentstart": "2026-04-03T08:00:00Z"},
			Raw:    map[string]any{"id": "b", "start": "2026-04-03T08:00:00Z", "end": "soon"},
		},
	})
}
