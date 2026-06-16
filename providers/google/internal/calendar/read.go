package calendar

import (
	"context"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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

	merged := mergeAndDedupeEvents(perCalendar)

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

	return rows, nil
}

// buildEventsURLForCalendar builds the first-page events URL for the given calendar.
// It mirrors the events-specific query params applied in buildReadURL.
func (a *Adapter) buildEventsURLForCalendar(
	calendarID string, params common.ReadParams,
) (*urlbuilder.URL, error) {
	objectPath, err := Schemas.FindURLPath(providers.ModuleGoogleCalendar, objectNameEvents)
	if err != nil {
		return nil, err
	}

	objectPath = strings.ReplaceAll(objectPath, "{calendarId}", calendarID)

	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, objectPath)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("maxResults", strconv.Itoa(defaultPageSize))

	// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
	if !params.Since.IsZero() {
		url.WithQueryParam("updatedMin", datautils.Time.FormatRFC3339inUTCWithMilliseconds(params.Since))
	}

	return url, nil
}

// mergeAndDedupeEvents flattens per-calendar event rows into one list, dropping
// duplicates that appear on more than one calendar. Insertion order is preserved.
func mergeAndDedupeEvents(perCalendar [][]common.ReadResultRow) []common.ReadResultRow {
	merged := make([]common.ReadResultRow, 0)
	seen := make(map[string]struct{})

	for _, rows := range perCalendar {
		for _, row := range rows {
			key := eventDedupeKey(row)
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

// eventDedupeKey identifies an event across calendars. It prefers iCalUID, which is
// stable for the same event on different calendars, and falls back to the event id.
func eventDedupeKey(row common.ReadResultRow) string {
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
