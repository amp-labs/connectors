package acculynx

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/amp-common/feature"
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"gotest.tools/v3/assert"
)

// TestReadAppointmentLookaheadRepeatedIDs covers calendars/appointments reads
// with appointmentLookaheadRepeatedIDsFlag on: the read cases of #3723, which
// this flag brings back, except that an occurrence on two calendars is now one
// row. It is not parallel because the flag is
// process-global: TestRead's parallel cases, which read it off, resume only
// after this test has returned and reset it.
//
//nolint:paralleltest,tparallel,funlen // flips the process-global appointment flags
func TestReadAppointmentLookaheadRepeatedIDs(t *testing.T) {
	assert.NilError(t, appointmentLookaheadRepeatedIDsFlag.Set(feature.On()))
	t.Cleanup(appointmentLookaheadRepeatedIDsFlag.Reset)

	tests := []testconn.TestCaseRead{
		{
			// AccuLynx rejects a range of 90 days or more, so a 120-day read is
			// two requests: 89 days, then the remainder.
			Name: "Repeated-id lookahead read splits a range of 90 days or more into windows",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: appointmentsServer(
				appointmentsWindowCase("2026-01-01", "2026-03-31",
					appointmentsPage(`{"id": "appt-1", "start": "2026-02-10T14:00:00Z"}`)),
				appointmentsWindowCase("2026-04-01", "2026-04-30",
					appointmentsPage(`{"id": "appt-2", "start": "2026-04-20T14:00:00Z"}`)),
			),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 2, Done: true},
		},
		{
			// A multi-day CrewLabor appointment comes back as one row per day, all
			// under one id: every day is kept, and the id repeats. The 03-31 row
			// is returned by both windows and is merged into one.
			Name: "Repeated-id lookahead read keeps each day of a multi-day appointment and drops window repeats",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id", "start"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: appointmentsServer(
				appointmentsWindowCase("2026-01-01", "2026-03-31",
					appointmentsPage(`{"id": "crew-1", "start": "2026-03-31T04:00:00Z", "allDay": true}`)),
				appointmentsWindowCase("2026-04-01", "2026-04-30",
					appointmentsPage(
						`{"id": "crew-1", "start": "2026-03-31T04:00:00Z", "allDay": true}`,
						`{"id": "crew-1", "start": "2026-04-01T04:00:00Z", "allDay": true}`,
						`{"id": "crew-1", "start": "2026-04-02T04:00:00Z", "allDay": true}`,
					)),
			),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Id:     "crew-1",
						Fields: map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z"},
						Raw:    map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z", "allDay": true},
					},
					{
						Id:     "crew-1",
						Fields: map[string]any{"id": "crew-1", "start": "2026-04-01T04:00:00Z"},
						Raw:    map[string]any{"id": "crew-1", "start": "2026-04-01T04:00:00Z", "allDay": true},
					},
					{
						Id:     "crew-1",
						Fields: map[string]any{"id": "crew-1", "start": "2026-04-02T04:00:00Z"},
						Raw:    map[string]any{"id": "crew-1", "start": "2026-04-02T04:00:00Z", "allDay": true},
					},
				},
				Done: true,
			},
		},
		{
			// The same occurrence on two calendars is merged into one row carrying
			// both users, so an id-and-start key never repeats in a read.
			Name: "Repeated-id lookahead read returns an occurrence on two calendars once, with both users",
			Input: common.ReadParams{
				ObjectName:        "calendars/appointments",
				Fields:            connectors.Fields("id"),
				Since:             time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Until:             time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
				AssociatedObjects: []string{"users"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/calendars"),
						},
						Then: mockserver.ResponseString(http.StatusOK, `{"count": 2, "pageSize": 25, "pageStartIndex": 0,
							"items": [{"id": "cal-001"}, {"id": "cal-002"}]}`),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/calendars/cal-001/appointments"),
						},
						Then: appointmentsPage(`{"id": "appt-1", "start": "2026-04-15T14:00:00Z"}`),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/calendars/cal-002/appointments"),
						},
						Then: appointmentsPage(`{"id": "appt-1", "start": "2026-04-15T14:00:00Z"}`),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id:     "appt-1",
						Fields: map[string]any{"id": "appt-1"},
						Raw:    map[string]any{"id": "appt-1"},
						Associations: map[string][]common.Association{
							"users": {{ObjectId: "cal-001"}, {ObjectId: "cal-002"}},
						},
					},
				},
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestReadConnector(tt.Server.URL)
			})
		})
	}
}

// TestAppointmentReadModeFor checks which flag decides a calendars/appointments
// read: plain by default, the occurrence flag over the repeated-id lookahead
// flag, and every other object plain whatever the flags.
//
//nolint:paralleltest // flips the process-global appointment flags
func TestAppointmentReadModeFor(t *testing.T) {
	t.Cleanup(appointmentLookaheadRepeatedIDsFlag.Reset)
	t.Cleanup(appointmentOccurrencesFlag.Reset)

	appointments := common.ReadParams{ObjectName: "calendars/appointments"}
	calendars := common.ReadParams{ObjectName: "calendars"}

	assert.Equal(t, appointmentReadModeFor(t.Context(), appointments), appointmentReadPlain, "off by default")

	assert.NilError(t, appointmentLookaheadRepeatedIDsFlag.Set(feature.On()))
	assert.Equal(t, appointmentReadModeFor(t.Context(), appointments), appointmentReadLookaheadRepeatedIDs, "lookahead on")

	assert.NilError(t, appointmentOccurrencesFlag.Set(feature.On()))
	assert.Equal(t, appointmentReadModeFor(t.Context(), appointments), appointmentReadOccurrences,
		"occurrences win over lookahead")
	assert.Equal(t, appointmentReadModeFor(t.Context(), calendars), appointmentReadPlain,
		"other objects are always plain")

	appointmentLookaheadRepeatedIDsFlag.Reset()
	assert.Equal(t, appointmentReadModeFor(t.Context(), appointments), appointmentReadOccurrences,
		"occurrences on their own")
}

// TestMergeRepeatedOccurrences pins the merge that both flagged reads use: rows
// with the same appointment id and start instant become one row with the
// associations of every repeat, ids unchanged and order kept, while every
// other occurrence, even on the same day, and every row without an id is
// kept. A row whose start does not parse merges only on its id and start text.
func TestMergeRepeatedOccurrences(t *testing.T) {
	t.Parallel()

	row := func(raw map[string]any, users ...string) common.ReadResultRow {
		result := common.ReadResultRow{Raw: raw}

		if id, ok := raw["id"].(string); ok {
			result.Id = id
		}

		for _, user := range users {
			if result.Associations == nil {
				result.Associations = map[string][]common.Association{}
			}

			result.Associations["users"] = append(result.Associations["users"], common.Association{ObjectId: user})
		}

		return result
	}

	rows := []common.ReadResultRow{
		row(map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z"}, "cal-1"),
		row(map[string]any{"title": "no id"}),
		row(map[string]any{"id": "crew-1", "start": "2026-04-01T04:00:00Z"}, "cal-1"),
		row(map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z"}, "cal-2"),
		row(map[string]any{"id": "crew-1", "start": "2026-03-30T23:00:00-05:00"}, "cal-1"),
		row(map[string]any{"id": "crew-1", "start": "2026-04-01T01:00:00Z"}, "cal-1"),
		row(map[string]any{"title": "no id"}),
		row(map[string]any{"id": "x"}, "cal-1"),
		row(map[string]any{"id": "x", "start": "soon"}, "cal-1"),
		row(map[string]any{"id": "x"}, "cal-2"),
	}

	assert.DeepEqual(t, mergeRepeatedOccurrences(rows), []common.ReadResultRow{
		row(map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z"}, "cal-1", "cal-2"),
		row(map[string]any{"title": "no id"}),
		row(map[string]any{"id": "crew-1", "start": "2026-04-01T04:00:00Z"}, "cal-1"),
		row(map[string]any{"id": "crew-1", "start": "2026-04-01T01:00:00Z"}, "cal-1"),
		row(map[string]any{"title": "no id"}),
		row(map[string]any{"id": "x"}, "cal-1", "cal-2"),
		row(map[string]any{"id": "x", "start": "soon"}, "cal-1"),
	})
}
