package acculynx

import (
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	errJobsWindowStalled = errors.New("acculynx: too many jobs share the same two-day modifiedDate span " +
		"to page past the pageStartIndex limit")
	errJobsNotOrdered = errors.New("acculynx: jobs page is not ordered by modifiedDate, " +
		"so the date window cannot be moved safely")
)

const (
	// maxPageStartIndex is the furthest position AccuLynx lets us read to.
	// If a page would end after position 100000 (pageStartIndex + pageSize >
	// 100000), AccuLynx returns 400 "Requested pageStartIndex is out of range.".
	// This happens no matter how many jobs exist, so one list can only give us
	// its first 100000 jobs.
	// Tested live on GET /jobs: start 99975 with pageSize 25 works, start 99976
	// with pageSize 25 fails.
	// https://apidocs.acculynx.com/reference/searchjobs
	maxPageStartIndex = 100000

	sortByParam         = "sortBy"
	sortOrderParam      = "sortOrder"
	dateFilterTypeParam = "dateFilterType"
	startDateParam      = "startDate"
	endDateParam        = "endDate"

	modifiedDateValue = "ModifiedDate"
	ascendingValue    = "Ascending"

	jobsModifiedDateField = "modifiedDate"
)

// How we read more than 100000 jobs
//
// Normally we keep AccuLynx's default order (by created date). A job's created
// date never changes, so jobs stay in the same place while we read, even if
// someone edits one. No job gets skipped.
//
// Sorting by modified date is risky. When someone edits a job, it moves to the
// end of the list, every job after it moves up one place, and the next page
// misses one job. So we only sort by modified date when we have to.
//
// We have to when the list has more than 100000 jobs. Then we:
//  1. Start the list again from position 0, sorted oldest-modified first.
//  2. Read pages as usual until the next page would go past 100000.
//  3. Ask for a new list that starts the day before the newest job we have
//     read, and start again from position 0 (advanceJobsWindow).
//
// AccuLynx filters by whole days, so a few jobs from that last day or two come
// twice. That is fine: getting a job twice is better than missing it.
//
// When the assigned jobs are done, the unassigned jobs start fresh with the
// read's own dates and the default order (restoreJobsWindowStart).
//
// All of these rules are in fitJobsCursorUnderCap. Every jobs next-page URL
// goes through it, whether we just built it or it was saved from an earlier
// read. Everything we need is stored in the URL itself.

// fitJobsCursorUnderCap changes a jobs next-page URL so AccuLynx will accept it.
// previousResponse is the response we got just before this URL, or nil when we
// do not have one (for example, a URL saved from an earlier read).
func (c *Connector) fitJobsCursorUnderCap(
	cursor *urlbuilder.URL,
	previousResponse *ajson.Node,
	params common.ReadParams,
) error {
	// Default order (by created date). We cannot move the start date forward in
	// this order, so the only way past 100000 is to sort by modified date and
	// start again from 0. We only do this when the list is too big, because the
	// default order is safer when someone edits a job while we read.
	if !hasJobsOrdering(cursor) {
		if listingPastOffsetCap(previousResponse, cursor) {
			restartJobsWindowOrdered(cursor)
		}

		return nil
	}

	// Sorted by modified date, and the next page still fits under 100000.
	if !pastOffsetCap(cursor) {
		return nil
	}

	// No previous response, so we do not know the newest date we have read.
	// Start this same list again from 0. When paging reaches 100000 again we
	// will have a response, and the date moves forward as usual.
	if previousResponse == nil {
		restartJobsWindowOrdered(cursor)

		return nil
	}

	// Start a new list from the newest date we have read.
	return advanceJobsWindow(cursor, previousResponse, c.arrayFieldName(objectJobs), params)
}

// pastOffsetCap says whether the next page would go past 100000
// (pageStartIndex + pageSize > 100000).
func pastOffsetCap(url *urlbuilder.URL) bool {
	start := queryParamIntOrDefault(url, pageStartParam, 0)
	size := queryParamIntOrDefault(url, pageSizeParam, maxPageSize)

	return start+size > maxPageStartIndex
}

// listingPastOffsetCap says whether paging cannot reach the end of the list.
func listingPastOffsetCap(previousResponse *ajson.Node, cursor *urlbuilder.URL) bool {
	// This is one way to know the list is too big: the next page would go past 100000.
	if pastOffsetCap(cursor) {
		return true
	}

	if previousResponse == nil {
		return false
	}

	count, _ := envelopeInt(previousResponse, countKey)

	// The other way to know: check the count field in the previous response.
	// If it is more than 100000, we cannot reach the end of the list.
	return count > maxPageStartIndex
}

// restartJobsWindowOrdered makes the URL start again from position 0 of the
// same list, sorted oldest-modified first.
func restartJobsWindowOrdered(cursor *urlbuilder.URL) {
	cursor.WithQueryParam(sortByParam, modifiedDateValue)
	cursor.WithQueryParam(sortOrderParam, ascendingValue)
	cursor.WithQueryParam(pageStartParam, "0")
}

// hasJobsOrdering says whether the URL asks for jobs sorted oldest-modified
// first (sortBy=ModifiedDate&sortOrder=Ascending).
//
// We need this order to move the start date forward. In this order, every job
// we have not read yet was modified on or after the newest job we have read,
// so a new list starting from that date skips nothing. In the default order
// (by created date) the modified dates are mixed up, so this is not safe.
func hasJobsOrdering(url *urlbuilder.URL) bool {
	sortBy, _ := url.GetFirstQueryParam(sortByParam)
	sortOrder, _ := url.GetFirstQueryParam(sortOrderParam)

	return sortBy == modifiedDateValue && sortOrder == ascendingValue
}

// makeJobsNextPage builds the next-page URL for jobs and passes it through
// fitJobsCursorUnderCap.
func (c *Connector) makeJobsNextPage(
	params common.ReadParams,
	reqURL *urlbuilder.URL,
	base common.NextPageFunc,
) common.NextPageFunc {
	return func(root *ajson.Node) (string, error) {
		next, err := base(root)
		if err != nil || next == "" || reqURL == nil {
			return next, err
		}

		nextURL, err := urlbuilder.New(next)
		if err != nil {
			return "", err
		}

		previousResponse := root

		// The assigned jobs are done and the unassigned jobs start. That is a new
		// list, so put back the read's own dates and the default order. Do not
		// pass the previous response: its job count belongs to the assigned list.
		if startsUnassignedSweep(reqURL, nextURL) {
			restoreJobsWindowStart(nextURL, params)

			previousResponse = nil
		}

		if err := c.fitJobsCursorUnderCap(nextURL, previousResponse, params); err != nil {
			return "", err
		}

		return nextURL.String(), nil
	}
}

func startsUnassignedSweep(reqURL, nextURL *urlbuilder.URL) bool {
	_, current := reqURL.GetFirstQueryParam(assignmentParam)
	_, next := nextURL.GetFirstQueryParam(assignmentParam)

	return !current && next
}

// restoreJobsWindowStart puts the read's own dates and the default order back
// on the URL.
func restoreJobsWindowStart(url *urlbuilder.URL, params common.ReadParams) {
	url.RemoveQueryParam(sortByParam)
	url.RemoveQueryParam(sortOrderParam)

	// The read had no Since/Until, so remove any date filter we added to get
	// past 100000.
	if params.Since.IsZero() && params.Until.IsZero() {
		url.RemoveQueryParam(dateFilterTypeParam)
		url.RemoveQueryParam(startDateParam)
		url.RemoveQueryParam(endDateParam)

		return
	}

	since, until := pairedDateWindow(params.Since, params.Until)
	url.WithQueryParam(dateFilterTypeParam, modifiedDateValue)
	url.WithQueryParam(startDateParam, since.Format(time.DateOnly))
	url.WithQueryParam(endDateParam, until.Format(time.DateOnly))
}

// advanceJobsWindow changes a URL that would go past 100000 into the first page
// of a new, shorter list that starts from the newest job we have read.
func advanceJobsWindow(
	cursor *urlbuilder.URL,
	previousResponse *ajson.Node,
	recordsKey string,
	params common.ReadParams,
) error {
	newest, err := newestModifiedDate(previousResponse, recordsKey)
	if err != nil {
		return err
	}

	// End date: keep the one already on the URL. A read without Since/Until has
	// none the first time it reaches 100000, so use the read's Until, or today.
	// AccuLynx needs both startDate and endDate.
	end := pairedUntilDate(params)

	if raw, ok := cursor.GetFirstQueryParam(endDateParam); ok {
		if parsed, parseErr := time.Parse(time.DateOnly, raw); parseErr == nil {
			end = parsed
		}
	}

	// Start date: the day before the newest job we have read. We go one day back
	// because AccuLynx filters by whole days and its day may not be in UTC. A
	// few jobs come twice, but none are missed.
	start := newest.UTC().Truncate(24*time.Hour).AddDate(0, 0, -1) //nolint:mnd
	// AccuLynx requires startDate to be earlier than endDate.
	if !start.Before(end) {
		start = end.AddDate(0, 0, -1)
	}

	// The start date of the list we are reading now.
	current := time.Unix(0, 0).UTC()

	if raw, ok := cursor.GetFirstQueryParam(startDateParam); ok {
		if parsed, parseErr := time.Parse(time.DateOnly, raw); parseErr == nil {
			current = parsed
		}
	}

	// The start date did not move forward: more than 100000 jobs were modified
	// within two days. Stop with an error instead of reading the same list
	// forever.
	if !start.After(current) {
		return fmt.Errorf("%w (startDate %s)", errJobsWindowStalled, current.Format(time.DateOnly))
	}

	// Ask for the new list, starting from position 0.
	cursor.WithQueryParam(dateFilterTypeParam, modifiedDateValue)
	cursor.WithQueryParam(startDateParam, start.Format(time.DateOnly))
	cursor.WithQueryParam(endDateParam, end.Format(time.DateOnly))
	cursor.WithQueryParam(pageStartParam, "0")

	return nil
}

// pairedUntilDate returns the read's Until, or today if there is none.
func pairedUntilDate(params common.ReadParams) time.Time {
	_, until := pairedDateWindow(params.Since, params.Until)

	// Cut off the time of day, so advanceJobsWindow compares whole days.
	return until.UTC().Truncate(24 * time.Hour) //nolint:mnd
}

// newestModifiedDate returns the modified date of the last job on the page.
func newestModifiedDate(root *ajson.Node, recordsKey string) (time.Time, error) {
	records, err := common.MakeRecordsFunc(recordsKey)(root)
	if err != nil {
		return time.Time{}, err
	}

	var newest time.Time

	// The URL already asks for this order (hasJobsOrdering checks it), and
	// AccuLynx returns 400 for bad sort values. This loop still checks the
	// order because the URL only shows what we asked for, not what we got.
	// If a page ever came back in another order, moving the start date
	// forward would skip jobs with no error. The loop costs little (a page
	// has at most 25 jobs) and turns that silent loss into an error.
	for _, record := range records {
		raw, err := jsonquery.New(record).StringRequired(jobsModifiedDateField)
		if err != nil {
			return time.Time{}, err
		}

		modified, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, fmt.Errorf("acculynx: job %s %q: %w", jobsModifiedDateField, raw, err)
		}

		if modified.Before(newest) {
			return time.Time{}, errJobsNotOrdered
		}

		newest = modified
	}

	// No jobs on the page, so there is no date to move forward from.
	if newest.IsZero() {
		return time.Time{}, errJobsNotOrdered
	}

	return newest, nil
}
