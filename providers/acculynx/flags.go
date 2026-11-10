package acculynx

import (
	"context"

	"github.com/amp-labs/amp-common/feature"
	"github.com/amp-labs/connectors/common"
)

// appointmentOccurrencesFlag turns on occurrence reads of calendars/appointments.
// AccuLynx returns a multi-day or recurring appointment as one row per scheduled
// day, all under the appointment's id. An occurrence read reaches 60 days ahead
// when Until is unset and gives each of those rows an id of its own (see
// appointmentOccurrences), so a read never repeats an id. Off, the read ends at
// Until or now and the rows are AccuLynx's as they come.
//
// Off by default. The server's read activity attaches the project, integration,
// installation, connection and provider to the context, so the rollout can
// target any of those (not the org).
var appointmentOccurrencesFlag = feature.Define("acculynx-appointment-occurrences", //nolint:gochecknoglobals
	feature.Off(),
	feature.Description("AccuLynx calendars/appointments reads reach 60 days ahead and return one row "+
		"per occurrence, with id <appointment id>_<UTC start date>; off, they end at Until or now and "+
		"repeat the appointment id for each day of a multi-day or recurring appointment"),
	feature.Owner("Chris Eberle"))

// readsAppointmentOccurrences reports whether a read of params is an occurrence
// read: the object is calendars/appointments and appointmentOccurrencesFlag is on
// for the subject in ctx.
func readsAppointmentOccurrences(ctx context.Context, params common.ReadParams) bool {
	return params.ObjectName == "calendars/appointments" && appointmentOccurrencesFlag.Enabled(ctx)
}

// appointmentMergedSpansFlag turns on merged appointment spans for
// calendars/appointments reads (see addMergedAppointmentSpans): the rows of
// each appointment, one per day or occurrence, are merged into one span, the
// earliest start and latest end among them, and every row of the appointment
// carries it as appointmentstart and appointmentend. The rows themselves are
// not merged and stay as they are. It is independent of
// appointmentOccurrencesFlag.
//
// Off by default; an experiment. The fields are not in the object's metadata.
var appointmentMergedSpansFlag = feature.Define("acculynx-appointment-merged-spans", //nolint:gochecknoglobals
	feature.Off(),
	feature.Description("AccuLynx calendars/appointments reads merge the rows of each appointment into one "+
		"span (earliest start, latest end) and add it to every row as appointmentstart and appointmentend; "+
		"the rows themselves stay as they are (experiment; the fields are not in the object's metadata)"),
	feature.Owner("Chris Eberle"))

// readsMergedAppointmentSpans reports whether a read of params adds merged
// appointment spans: the object is calendars/appointments and
// appointmentMergedSpansFlag is on for the subject in ctx.
func readsMergedAppointmentSpans(ctx context.Context, params common.ReadParams) bool {
	return params.ObjectName == "calendars/appointments" && appointmentMergedSpansFlag.Enabled(ctx)
}

// appointmentLookaheadRepeatedIDsFlag turns on repeated-id lookahead reads of
// calendars/appointments, the behavior of #3723 before it was reverted: a read
// reaches 60 days ahead when Until is unset, in windows of at most 89 days, and
// keeps AccuLynx's ids. A multi-day or recurring appointment comes back as one
// row per day, all under its id, so the read repeats that id once per day. It
// drops only a row that two windows of one calendar both return (same id and
// start; see fetchAppointmentLookaheadRepeatedIDs). Off, the read ends at Until
// or now.
//
// Off by default. appointmentOccurrencesFlag takes precedence when both are on
// (see appointmentReadModeFor).
//
//nolint:gochecknoglobals
var appointmentLookaheadRepeatedIDsFlag = feature.Define("acculynx-appointment-lookahead-repeated-ids",
	feature.Off(),
	feature.Description("AccuLynx calendars/appointments reads reach 60 days ahead and repeat AccuLynx's id "+
		"once per day of a multi-day or recurring appointment; acculynx-appointment-occurrences takes "+
		"precedence when both are on"),
	feature.Owner("Chris Eberle"))

// readsAppointmentLookaheadRepeatedIDs reports whether
// appointmentLookaheadRepeatedIDsFlag is on for a calendars/appointments read
// of params and the subject in ctx. On its own it does not decide the read; see
// appointmentReadModeFor.
func readsAppointmentLookaheadRepeatedIDs(ctx context.Context, params common.ReadParams) bool {
	return params.ObjectName == "calendars/appointments" && appointmentLookaheadRepeatedIDsFlag.Enabled(ctx)
}

// appointmentReadMode is how a calendars/appointments read gets its rows.
type appointmentReadMode int

const (
	// appointmentReadPlain reads one date range, Since to Until or now, and
	// returns the rows as AccuLynx returns them (see applyAppointmentsDateWindow).
	// Every read of another object is plain too.
	appointmentReadPlain appointmentReadMode = iota
	// appointmentReadLookaheadRepeatedIDs is a repeated-id lookahead read (see
	// appointmentLookaheadRepeatedIDsFlag).
	appointmentReadLookaheadRepeatedIDs
	// appointmentReadOccurrences is an occurrence read (see
	// appointmentOccurrencesFlag).
	appointmentReadOccurrences
)

// appointmentReadModeFor returns the read mode of params for the subject in
// ctx. appointmentOccurrencesFlag wins over appointmentLookaheadRepeatedIDsFlag:
// both read 60 days ahead, and only an occurrence read keeps a read from
// repeating an id.
func appointmentReadModeFor(ctx context.Context, params common.ReadParams) appointmentReadMode {
	switch {
	case readsAppointmentOccurrences(ctx, params):
		return appointmentReadOccurrences
	case readsAppointmentLookaheadRepeatedIDs(ctx, params):
		return appointmentReadLookaheadRepeatedIDs
	default:
		return appointmentReadPlain
	}
}
