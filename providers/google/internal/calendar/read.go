package calendar

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/google/internal/core"
)

// maxReadConcurrency bounds how many calendars we read events from at once.
//
// The Calendar API allows 600 queries per minute per user (10 qps) and rate limits
// bursts via a sliding window, so we keep parallelism well below that ceiling to
// avoid 403/429 usageLimits responses.
// https://developers.google.com/workspace/calendar/api/guides/quota
const maxReadConcurrency = 5

// Read overrides the default HTTP reader. When ReadParams.Opts requests events from
// all calendars it reads across the whole calendar list and merges the results;
// otherwise it delegates to the standard reader, which reads only the primary calendar.
func (a *Adapter) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if !shouldReadEventsForAllCalendars(params) {
		return a.Reader.Read(ctx, params)
	}

	return a.readEventsAcrossCalendars(ctx, params)
}

// shouldReadEventsForAllCalendars reports whether params opts into reading events from
// every calendar. It is false unless the object is "events" and ReadParams.Opts asserts
// to ReadParamsOpts with the flag set; an empty or mismatched Opts keeps default behavior.
func shouldReadEventsForAllCalendars(params common.ReadParams) bool {
	if params.ObjectName != objectNameEvents {
		return false
	}

	opts, ok := params.Opts.(core.ReadParamsOpts)
	if !ok {
		return false
	}

	return opts.ReadEventsForAllCalendars
}

// readEventsAcrossCalendars lists every calendar, reads each one's events concurrently
// (bounded by maxReadConcurrency), then merges them into a single deduplicated result.
func (a *Adapter) readEventsAcrossCalendars(
	ctx context.Context, params common.ReadParams,
) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	calendarIDs, err := a.listAllCalendarIDs(ctx)
	if err != nil {
		return nil, err
	}

	// Each job writes only into its own slice index, so no locking is required.
	perCalendar := make([][]common.ReadResultRow, len(calendarIDs))
	jobs := make([]simultaneously.Job, len(calendarIDs))

	for i, calendarID := range calendarIDs {
		jobs[i] = func(ctx context.Context) error {
			rows, err := a.readAllEventPages(ctx, calendarID, params)
			if err != nil {
				return err
			}

			perCalendar[i] = rows

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxReadConcurrency, jobs...); err != nil {
		return nil, err
	}

	opts, _ := params.Opts.(core.ReadParamsOpts)
	merged := mergeAndDedupeEvents(perCalendar, opts.SingleEvents)

	return &common.ReadResult{
		Rows:     int64(len(merged)),
		Data:     merged,
		NextPage: "",
		Done:     true,
	}, nil
}

// listAllCalendarIDs drains the calendarList object and returns every calendar's ID.
func (a *Adapter) listAllCalendarIDs(ctx context.Context) ([]string, error) {
	listParams := common.ReadParams{
		ObjectName: objectNameCalendarList,
		Fields:     datautils.NewStringSet("id"),
	}

	var ids []string

	for {
		result, err := a.Reader.Read(ctx, listParams)
		if err != nil {
			return nil, err
		}

		for _, row := range result.Data {
			if id := rawString(row, "id"); id != "" {
				ids = append(ids, id)
			}
		}

		if result.Done || result.NextPage == "" {
			break
		}

		listParams.NextPage = result.NextPage
	}

	return ids, nil
}

// readAllEventPages drains every page of events for a single calendar.
//
// It reuses the standard HTTP reader by seeding ReadParams.NextPage with the
// per-calendar events URL; buildReadURL returns NextPage verbatim, so the reader
// (and its pagination) stays scoped to this calendar.
func (a *Adapter) readAllEventPages(
	ctx context.Context, calendarID string, params common.ReadParams,
) ([]common.ReadResultRow, error) {
	url, err := a.buildEventsURLForCalendar(calendarID, params)
	if err != nil {
		return nil, err
	}

	pageParams := params
	pageParams.NextPage = common.NextPageToken(url.String())

	var rows []common.ReadResultRow

	for {
		result, err := a.Reader.Read(ctx, pageParams)
		if err != nil {
			return nil, err
		}

		rows = append(rows, result.Data...)

		if result.Done || result.NextPage == "" {
			break
		}

		pageParams.NextPage = result.NextPage
	}

	// singleEvents expansion omits the series master; fetch it separately when requested.
	masters, err := a.fetchSeriesMasters(ctx, calendarID, rows, params)
	if err != nil {
		return nil, err
	}

	return append(rows, masters...), nil
}

// fetchSeriesMasters fetches the recurring-event master for each distinct recurringEventId
// found among the instance rows and returns them as additional rows. It is a no-op unless
// FetchSeriesMasters is set, because events.list with singleEvents=true expands recurring
// events into instances and omits the series master (events.list never returns both).
//
// Masters are fetched from the same calendar the instances came from (an event id is scoped
// to its calendar), concurrency-bounded like the per-calendar reads. A master that no longer
// exists (404) is skipped rather than failing the whole read — its instances may still be
// present as cancelled rows when showDeleted is set.
//
// https://developers.google.com/workspace/calendar/api/v3/reference/events/get
func (a *Adapter) fetchSeriesMasters(
	ctx context.Context, calendarID string, instances []common.ReadResultRow, params common.ReadParams,
) ([]common.ReadResultRow, error) {
	opts, _ := params.Opts.(core.ReadParamsOpts)
	if !opts.FetchSeriesMasters {
		return nil, nil
	}

	masterIDs := distinctRecurringEventIDs(instances)
	if len(masterIDs) == 0 {
		return nil, nil
	}

	// Each job writes only into its own slice index, so no locking is required. A nil entry
	// marks a master that was skipped (not found).
	fetched := make([]*common.ReadResultRow, len(masterIDs))
	jobs := make([]simultaneously.Job, len(masterIDs))

	for idx, masterID := range masterIDs {
		jobs[idx] = func(ctx context.Context) error {
			row, err := a.fetchEvent(ctx, calendarID, masterID, params.Fields)
			if err != nil {
				// A series master can be legitimately missing on the calendar we read
				// the instance from: e.g. a shared recurring meeting whose master lives
				// on the organizer's calendar (an attendee's copy 404s on that id), or a
				// series whose anchor occurrence has been cancelled. Google returns 404
				// for such a get. Skip that one master instead of failing the whole read
				// — its instances are already included (and present as cancelled rows
				// when showDeleted is set).
				if isNotFound(err) {
					logging.Logger(ctx).Warn(
						"google calendar: recurring series master not found, skipping",
						"calendarId", calendarID,
						"masterId", masterID,
					)

					return nil
				}

				return err
			}

			fetched[idx] = &row

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxReadConcurrency, jobs...); err != nil {
		return nil, err
	}

	masters := make([]common.ReadResultRow, 0, len(fetched))

	for _, row := range fetched {
		if row != nil {
			masters = append(masters, *row)
		}
	}

	return masters, nil
}

// isNotFound reports whether err represents an HTTP 404 from the Calendar API.
//
// fetchEvent calls the base JSON HTTP client directly (bypassing the reader), and that
// client's default error handler maps 404 to a retryable error rather than
// common.ErrNotFound. So, in addition to the sentinel, we inspect the HTTP status on the
// wrapped *common.HTTPError — otherwise a genuinely-missing master would be treated as a
// transient failure and retried forever, wedging the whole read.
func isNotFound(err error) bool {
	if errors.Is(err, common.ErrNotFound) {
		return true
	}

	if httpErr, ok := errors.AsType[*common.HTTPError](err); ok {
		return httpErr.Status == http.StatusNotFound
	}

	return false
}

// distinctRecurringEventIDs returns the unique, non-empty recurringEventId values among the
// rows. For an instance of a recurring event this field is the id of the series master, so
// the result is the set of master ids to fetch (one per series, not one per instance).
func distinctRecurringEventIDs(rows []common.ReadResultRow) []string {
	seen := make(map[string]struct{})
	ids := make([]string, 0, len(rows))

	for _, row := range rows {
		id := rawString(row, "recurringEventId")
		if id == "" {
			continue
		}

		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	return ids
}

// fetchEvent fetches a single event by id (events.get) and returns it shaped like the list
// rows — the same selected Fields plus the full Raw payload — so it merges and dedupes
// consistently with instance rows.
func (a *Adapter) fetchEvent(
	ctx context.Context, calendarID, eventID string, fields datautils.StringSet,
) (common.ReadResultRow, error) {
	url, err := a.eventsBaseURL(calendarID)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	url.AddPath(eventID)

	resp, err := a.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return common.ReadResultRow{}, err
	}

	body, ok := resp.Body()
	if !ok {
		return common.ReadResultRow{}, common.ErrEmptyJSONHTTPResponse
	}

	record, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	rows, err := common.GetMarshaledData([]map[string]any{record}, fields.List())
	if err != nil {
		return common.ReadResultRow{}, err
	}

	return rows[0], nil
}

// buildEventsURLForCalendar builds the first-page events URL for the given calendar.
// It mirrors the events-specific query params applied in buildReadURL.
func (a *Adapter) buildEventsURLForCalendar(
	calendarID string, params common.ReadParams,
) (*urlbuilder.URL, error) {
	url, err := a.eventsBaseURL(calendarID)
	if err != nil {
		return nil, err
	}

	applyEventsQueryParams(url, params)

	return url, nil
}

// eventsBaseURL builds the events collection URL for a specific calendar id (no query params).
func (a *Adapter) eventsBaseURL(calendarID string) (*urlbuilder.URL, error) {
	objectPath, err := Schemas.FindURLPath(providers.ModuleGoogleCalendar, objectNameEvents)
	if err != nil {
		return nil, err
	}

	objectPath = strings.ReplaceAll(objectPath, "{calendarId}", calendarID)

	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, objectPath)
}

// mergeAndDedupeEvents flattens per-calendar event rows into one list, dropping
// duplicates that appear on more than one calendar. Insertion order is preserved.
func mergeAndDedupeEvents(perCalendar [][]common.ReadResultRow, singleEvents bool) []common.ReadResultRow {
	merged := make([]common.ReadResultRow, 0)
	seen := make(map[string]struct{})

	for _, rows := range perCalendar {
		for _, row := range rows {
			key := eventDedupeKey(row, singleEvents)
			if key != "" {
				if _, exists := seen[key]; exists {
					continue
				}

				seen[key] = struct{}{}
			}

			merged = append(merged, row)
		}
	}

	return merged
}

// eventDedupeKey identifies an event across calendars.
//
// With singleEvents expansion every instance of one recurring series shares a single iCalUID,
// so we must key on the per-instance event id to keep the instances distinct (and to keep the
// separately fetched series master, which has its own id). Otherwise we prefer iCalUID, which
// is stable for the same event copied onto multiple calendars, and fall back to the event id.
func eventDedupeKey(row common.ReadResultRow, singleEvents bool) string {
	if singleEvents {
		return rawString(row, "id")
	}

	if uid := rawString(row, "iCalUID"); uid != "" {
		return uid
	}

	return rawString(row, "id")
}

// rawString returns the string value at key in the row's raw payload, or "".
func rawString(row common.ReadResultRow, key string) string {
	if value, ok := row.Raw[key].(string); ok {
		return value
	}

	return ""
}
