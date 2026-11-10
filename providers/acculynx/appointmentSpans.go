package acculynx

import (
	"time"

	"github.com/amp-labs/connectors/common"
)

// The fields addAppointmentSpans sets, lowercase like every other field of a
// row (see common.ExtractLowercaseFieldsFromRaw).
const (
	appointmentStartField = "appointmentstart"
	appointmentEndField   = "appointmentend"
)

// appointmentSpan is the extent of one appointment within a read: the earliest
// start and the latest end among its rows, compared as instants and kept as
// AccuLynx wrote them.
type appointmentSpan struct {
	start, end         time.Time
	startText, endText string
}

// addAppointmentSpans sets appointmentstart and appointmentend on every row of
// a calendars/appointments read: the earliest start and the latest end among
// the rows that share the row's AccuLynx appointment id (Raw "id"). The rows
// stay as they are, one per day or occurrence; only the two fields are added.
//
// The span is the appointment's extent within this read, not its definition:
// it is clipped to the read's date range, rows are only grouped within one page
// of the calendar list, and for a recurring appointment it says nothing about
// the repeat pattern.
//
// It runs only when appointmentSpansFlag is on (see readsAppointmentSpans). The
// fields are not in the object's metadata, so they are set whether or not they
// were requested. Rows without an id get neither field, and a field is left out
// when no row of the appointment has an RFC 3339 value for it.
func addAppointmentSpans(rows []common.ReadResultRow) {
	spans := appointmentSpans(rows)

	for idx := range rows {
		appointmentID, _ := rows[idx].Raw["id"].(string)

		span, ok := spans[appointmentID]
		if !ok {
			continue
		}

		if rows[idx].Fields == nil {
			rows[idx].Fields = make(map[string]any)
		}

		if span.startText != "" {
			rows[idx].Fields[appointmentStartField] = span.startText
		}

		if span.endText != "" {
			rows[idx].Fields[appointmentEndField] = span.endText
		}
	}
}

// appointmentSpans returns the span of every appointment id among rows.
func appointmentSpans(rows []common.ReadResultRow) map[string]appointmentSpan {
	spans := make(map[string]appointmentSpan)

	for _, row := range rows {
		appointmentID, _ := row.Raw["id"].(string)
		if appointmentID == "" {
			continue
		}

		span := spans[appointmentID]

		if start, text, ok := appointmentTime(row, "start"); ok && (span.startText == "" || start.Before(span.start)) {
			span.start, span.startText = start, text
		}

		if end, text, ok := appointmentTime(row, "end"); ok && (span.endText == "" || end.After(span.end)) {
			span.end, span.endText = end, text
		}

		spans[appointmentID] = span
	}

	return spans
}

// appointmentTime returns the row's Raw value for key as an instant and as
// AccuLynx wrote it. It reports false when the value is not an RFC 3339 time.
func appointmentTime(row common.ReadResultRow, key string) (time.Time, string, bool) {
	parsed, ok := parseAppointmentTime(row.Raw[key])
	if !ok {
		return time.Time{}, "", false
	}

	text, _ := row.Raw[key].(string)

	return parsed, text, true
}
