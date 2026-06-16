package core

// ReadParamsOpts are Google-specific options for common.ReadParams.Opts.
//
// It lives in this shared leaf package so every Google module (calendar, mail,
// contacts, ...) can assert it without importing the top-level google package,
// which would create an import cycle. The google package re-exports it as
// google.ReadParamsOpts for external callers.
type ReadParamsOpts struct {
	// ReadEventsForAllCalendars, when true, reads "events" from every calendar in the
	// user's calendar list and merges the results into a single deduplicated list,
	// instead of reading only the primary calendar. It has no effect on other objects.
	ReadEventsForAllCalendars bool
}
