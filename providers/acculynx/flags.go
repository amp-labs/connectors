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

// appointmentSpansFlag turns on appointment spans on calendars/appointments
// rows (see addAppointmentSpans): every row carries appointmentstart and
// appointmentend, the earliest start and latest end among the read's rows of
// the same appointment. The rows themselves are unchanged. It is independent
// of appointmentOccurrencesFlag.
//
// Off by default; an experiment. The fields are not in the object's metadata.
var appointmentSpansFlag = feature.Define("acculynx-appointment-spans", //nolint:gochecknoglobals
	feature.Off(),
	feature.Description("AccuLynx calendars/appointments rows carry appointmentstart and appointmentend, "+
		"the earliest start and latest end among the read's rows with the same appointment id "+
		"(experiment; the fields are not in the object's metadata)"),
	feature.Owner("Chris Eberle"))

// readsAppointmentSpans reports whether a read of params adds appointment spans:
// the object is calendars/appointments and appointmentSpansFlag is on for the
// subject in ctx.
func readsAppointmentSpans(ctx context.Context, params common.ReadParams) bool {
	return params.ObjectName == "calendars/appointments" && appointmentSpansFlag.Enabled(ctx)
}
