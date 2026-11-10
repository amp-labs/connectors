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
