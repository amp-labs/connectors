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
