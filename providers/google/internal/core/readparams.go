package core

import "time"

// ReadParamsOpts are Google-specific options for common.ReadParams.Opts.
//
// It lives in this shared leaf package so every Google module (calendar, mail,
// contacts, ...) can assert it without importing the top-level google package,
// which would create an import cycle. The google package re-exports it as
// google.ReadParamsOpts for external callers.
//
// All fields below other than ReadEventsForAllCalendars tune the Google Calendar
// "events" read; the calendar adapter maps them onto events.list query params (and
// the series-master fan-out). They have no effect on other objects.
// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
type ReadParamsOpts struct {
	// ReadEventsForAllCalendars, when true, reads "events" from every calendar in the
	// user's calendar list and merges the results into a single deduplicated list,
	// instead of reading only the primary calendar. It has no effect on other objects.
	ReadEventsForAllCalendars bool

	// TimeMin and TimeMax bound the events.list window by event start time. Zero values
	// are omitted, leaving the API default (unbounded).
	TimeMin time.Time
	TimeMax time.Time

	// SingleEvents, when true, expands recurring events into individual instances
	// (events.list singleEvents=true). The series master is not returned by the list in
	// this mode; set FetchSeriesMasters to additionally fetch it.
	SingleEvents bool

	// ShowDeleted, when true, includes deleted/cancelled events (events.list showDeleted=true).
	ShowDeleted bool

	// MaxResults overrides the per-page size for events.list. Zero keeps the default.
	MaxResults int

	// EventTypes restricts the event types returned (events.list eventTypes, repeated). Empty keeps the default.
	EventTypes []string

	// OrderBy sets the events.list ordering ("startTime" or "updated"). Empty keeps the default.
	// "startTime" is only valid together with SingleEvents.
	OrderBy string

	// FetchSeriesMasters, when true, additionally fetches each recurring series' master event
	// (events.get on an instance's recurringEventId) and includes it in the results. This is
	// only meaningful with SingleEvents, which expands instances and omits the series master.
	// It is honored on the read-all-calendars path.
	FetchSeriesMasters bool
}
