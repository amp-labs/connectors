package acculynx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

// fakeJob is one synthetic AccuLynx job.
type fakeJob struct {
	id       string
	created  int // creation order, the API's default sort
	modified time.Time
}

// fakeJobsAPI mimics GET /api/v2/jobs closely enough to reproduce the
// production failure (ENG-4388): 25 jobs per page at most, a hard 400 at
// pageStartIndex >= 100000, two disjoint assignment listings, day-level
// startDate/endDate filtering evaluated in a given time zone, CreatedDate order
// by default and ModifiedDate order on request.
type fakeJobsAPI struct {
	assigned   []fakeJob
	unassigned []fakeJob
	location   *time.Location

	// editEvery > 0 edits one already delivered job every editEvery requests:
	// the first job of the listing gets modifiedDate editTo, as if a user
	// changed it while the read was paging.
	editEvery int
	editTo    time.Time

	mu             sync.Mutex
	requests       int
	maxStartIndex  int
	outOfRangeHits int
	cache          map[string][]fakeJob
}

func newFakeJobsAPI(assigned, unassigned []fakeJob, location *time.Location) *fakeJobsAPI {
	return &fakeJobsAPI{
		assigned:   assigned,
		unassigned: unassigned,
		location:   location,
		cache:      map[string][]fakeJob{},
	}
}

func (f *fakeJobsAPI) server(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v2/company-settings/custom-fields":
			_, _ = w.Write(customFieldDefinitionsEmptyResponse) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		case "/api/v2/jobs":
			f.serveJobs(w, r.URL.Query())
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

func writeFakeError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func (f *fakeJobsAPI) serveJobs(w http.ResponseWriter, query url.Values) {
	pageSize, start := 10, 0

	if raw := query.Get("pageSize"); raw != "" {
		pageSize, _ = strconv.Atoi(raw)
	}

	if raw := query.Get("pageStartIndex"); raw != "" {
		start, _ = strconv.Atoi(raw)
	}

	f.mu.Lock()
	f.requests++
	f.maxStartIndex = max(f.maxStartIndex, start)
	editNow := f.editEvery > 0 && start > 0 && f.requests%f.editEvery == 0
	f.mu.Unlock()

	if pageSize <= 0 || pageSize > 25 {
		writeFakeError(w, "pageSize must be > 0 and <= 25.")

		return
	}

	// Live AccuLynx rejects any page that would end past the ceiling, e.g.
	// pageStartIndex=99976&pageSize=25 or 99999&pageSize=2.
	if start+pageSize > maxPageStartIndex {
		f.mu.Lock()
		f.outOfRangeHits++
		f.mu.Unlock()
		writeFakeError(w, "Requested pageStartIndex is out of range.")

		return
	}

	list, problem := f.filtered(query)
	if problem != "" {
		writeFakeError(w, problem)

		return
	}

	if editNow && len(list) > 0 {
		f.edit(list[0].id)
		list, _ = f.filtered(query)
	}

	end := min(start+pageSize, len(list))
	items := make([]map[string]any, 0, max(end-start, 0))

	for _, job := range list[min(start, len(list)):end] {
		items = append(items, map[string]any{
			"id":               job.id,
			"currentMilestone": "Prospect",
			"modifiedDate":     job.modified.Format(time.RFC3339),
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"count":          len(list),
		"pageSize":       pageSize,
		"pageStartIndex": start,
		"items":          items,
	})
}

func (f *fakeJobsAPI) edit(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, list := range [][]fakeJob{f.assigned, f.unassigned} {
		for i := range list {
			if list[i].id == id {
				list[i].modified = f.editTo
			}
		}
	}

	clear(f.cache)
}

// filtered returns the listing the query asks for, cached per query shape.
func (f *fakeJobsAPI) filtered(query url.Values) ([]fakeJob, string) {
	startDate, endDate := query.Get("startDate"), query.Get("endDate")
	if (startDate == "") != (endDate == "") {
		return nil, "Start Date and End Date do not have the same format."
	}

	key := fmt.Sprintf("%s|%s|%s|%s|%s",
		query.Get("assignment"), startDate, endDate, query.Get("sortBy"), query.Get("sortOrder"))

	f.mu.Lock()
	defer f.mu.Unlock()

	if cached, ok := f.cache[key]; ok {
		return cached, ""
	}

	source := f.assigned
	if query.Get("assignment") == "unassigned" {
		source = f.unassigned
	}

	list := make([]fakeJob, 0, len(source))

	if startDate == "" {
		list = append(list, source...)
	} else {
		from, err1 := time.ParseInLocation(time.DateOnly, startDate, f.location)
		to, err2 := time.ParseInLocation(time.DateOnly, endDate, f.location)

		if err1 != nil || err2 != nil {
			return nil, "Invalid date."
		}

		if !from.Before(to) {
			return nil, "startDate must be earlier than endDate."
		}

		to = to.AddDate(0, 0, 1) // endDate is inclusive

		for _, job := range source {
			if !job.modified.Before(from) && job.modified.Before(to) {
				list = append(list, job)
			}
		}
	}

	if query.Get("sortBy") == "ModifiedDate" && query.Get("sortOrder") == "Ascending" {
		sort.SliceStable(list, func(i, j int) bool { return list[i].modified.Before(list[j].modified) })
	} else {
		sort.SliceStable(list, func(i, j int) bool { return list[i].created < list[j].created })
	}

	f.cache[key] = list

	return list, ""
}

// makeFakeJobs spreads count jobs evenly over [from, to). Creation order is
// scrambled relative to modifiedDate, as in a real account.
func makeFakeJobs(prefix string, count int, from, to time.Time) []fakeJob {
	jobs := make([]fakeJob, count)
	step := to.Sub(from) / time.Duration(count)

	for i := range jobs {
		jobs[i] = fakeJob{
			id:       prefix + strconv.Itoa(i),
			created:  (i * 7919) % count, // 7919 is prime: a fixed permutation
			modified: from.Add(time.Duration(i) * step),
		}
	}

	return jobs
}

func idSet(jobs ...[]fakeJob) map[string]bool {
	set := map[string]bool{}

	for _, list := range jobs {
		for _, job := range list {
			set[job.id] = true
		}
	}

	return set
}

func inRange(jobs []fakeJob, from, to time.Time) []fakeJob {
	var out []fakeJob

	for _, job := range jobs {
		if !job.modified.Before(from) && job.modified.Before(to) {
			out = append(out, job)
		}
	}

	return out
}

type readOutcome struct {
	seen       map[string]int
	rows       int
	reads      int
	err        error
	done       bool
	duplicates int
}

// readAll drives Read the way the Ampersand server does: page after page,
// handing back NextPage, until Done or an error.
func readAll(t *testing.T, serverURL string, params common.ReadParams) readOutcome {
	t.Helper()

	conn, err := constructTestConnector(serverURL)
	assert.NilError(t, err)

	outcome := readOutcome{seen: map[string]int{}}

	for outcome.reads < 50000 {
		outcome.reads++

		result, err := conn.Read(t.Context(), params)
		if err != nil {
			outcome.err = err

			return outcome
		}

		for _, row := range result.Data {
			id, _ := row.Raw["id"].(string)
			outcome.seen[id]++
			outcome.rows++
		}

		if result.Done {
			outcome.done = true

			break
		}

		params.NextPage = result.NextPage
	}

	for _, n := range outcome.seen {
		outcome.duplicates += n - 1
	}

	return outcome
}

func assertAllSeen(t *testing.T, outcome readOutcome, want map[string]bool) {
	t.Helper()

	missing := 0

	for id := range want {
		if outcome.seen[id] == 0 {
			missing++
		}
	}

	assert.Equal(t, missing, 0, "jobs never delivered: %d of %d", missing, len(want))
}

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

type jobsWindowCase struct {
	name       string
	assigned   []fakeJob
	unassigned []fakeJob
	location   *time.Location
	since      time.Time
	until      time.Time
	cursor     string // saved next-page query, appended to the server URL
	pageSize   int
	editEvery  int
	editTo     time.Time
	want       map[string]bool
	wantErr    error
	noDupes    bool
}

// TestJobsReadPastOffsetCap drives full reads against fakeJobsAPI. Before the
// fix the first two cases failed exactly like production (ENG-4388):
// HTTP 400 "Requested pageStartIndex is out of range.".
func TestJobsReadPastOffsetCap(t *testing.T) { //nolint:funlen
	t.Parallel()

	chicago, err := time.LoadLocation("America/Chicago")
	assert.NilError(t, err)

	backfillSince, backfillUntil := date(2026, 7, 3), date(2026, 10, 1)
	stuckCursor := "/api/v2/jobs?dateFilterType=ModifiedDate&endDate=2026-10-01&includes=initialAppointment" +
		"&pageSize=25&pageStartIndex=100000&startDate=2026-07-03"
	inFlightCursor := "/api/v2/jobs?dateFilterType=ModifiedDate&endDate=2026-10-01&includes=initialAppointment" +
		"&pageSize=25&pageStartIndex=99900&startDate=2026-07-03"
	now := time.Now().UTC()
	busyHour := date(2026, 8, 10).Add(15 * time.Hour)

	// The Hatch-like backfill: jobs spill a few days past both bounds so the
	// date filter matters; the connector-side filter trims to exact bounds.
	backfillAssigned := makeFakeJobs("a", 230000, backfillSince.AddDate(0, 0, -5), backfillUntil.AddDate(0, 0, 5))
	backfillUnassigned := makeFakeJobs("u", 30000, backfillSince.AddDate(0, 0, -5), backfillUntil.AddDate(0, 0, 5))

	stuckAssigned := makeFakeJobs("a", 160000, backfillSince, backfillUntil)
	stuckUnassigned := makeFakeJobs("u", 5000, backfillSince, backfillUntil)

	inFlightAssigned := makeFakeJobs("a", 130000, backfillSince, backfillUntil)
	inFlightUnassigned := makeFakeJobs("u", 2000, backfillSince, backfillUntil)

	fullAssigned := makeFakeJobs("a", 120000, now.AddDate(-3, 0, 0), now.Add(-time.Hour))
	fullUnassigned := makeFakeJobs("u", 3000, now.AddDate(-3, 0, 0), now.Add(-time.Hour))

	// Data well inside the bounds, so only the window move is under test.
	localAssigned := makeFakeJobs("a", 230000, backfillSince.AddDate(0, 0, 3), backfillUntil.AddDate(0, 0, -3))
	localUnassigned := makeFakeJobs("u", 4000, backfillSince.AddDate(0, 0, 3), backfillUntil.AddDate(0, 0, -3))

	smallAssigned := makeFakeJobs("a", 4000, date(2026, 10, 1), date(2026, 10, 8))
	smallUnassigned := makeFakeJobs("u", 300, date(2026, 10, 1), date(2026, 10, 8))

	editedAssigned := makeFakeJobs("a", 2000, date(2026, 10, 1), date(2026, 10, 8))

	tests := []jobsWindowCase{
		{
			name:       "Backfill with more than 100000 assigned jobs crosses the cap",
			assigned:   backfillAssigned,
			unassigned: backfillUnassigned,
			since:      backfillSince,
			until:      backfillUntil,
			want: idSet(inRange(backfillAssigned, backfillSince, backfillUntil),
				inRange(backfillUnassigned, backfillSince, backfillUntil)),
		},
		{
			name:       "Cursor saved at pageStartIndex=100000 recovers",
			assigned:   stuckAssigned,
			unassigned: stuckUnassigned,
			since:      backfillSince,
			until:      backfillUntil,
			cursor:     stuckCursor,
			want:       idSet(stuckAssigned, stuckUnassigned),
		},
		{
			name:       "Unordered cursor already in flight re-reads its window ordered at the cap",
			assigned:   inFlightAssigned,
			unassigned: inFlightUnassigned,
			since:      backfillSince,
			until:      backfillUntil,
			cursor:     inFlightCursor,
			want:       idSet(inFlightAssigned, inFlightUnassigned),
		},
		{
			name:       "Read without Since/Until adds a date filter only to cross the cap",
			assigned:   fullAssigned,
			unassigned: fullUnassigned,
			want:       idSet(fullAssigned, fullUnassigned),
		},
		{
			name:       "Window move leaves no gap when the server's day is not UTC",
			assigned:   localAssigned,
			unassigned: localUnassigned,
			location:   chicago,
			since:      backfillSince,
			until:      backfillUntil,
			want:       idSet(localAssigned, localUnassigned),
		},
		{
			name:       "Small read is unchanged and has no duplicates",
			assigned:   smallAssigned,
			unassigned: smallUnassigned,
			since:      date(2026, 10, 1),
			until:      date(2026, 10, 8),
			want:       idSet(smallAssigned, smallUnassigned),
			noDupes:    true,
		},
		{
			// Page sizes that do not divide 100000 must stop short of the
			// ceiling too: 99999&pageSize=3 is already out of range.
			name:       "Page size that does not divide the ceiling crosses it",
			assigned:   fullAssigned,
			unassigned: fullUnassigned,
			pageSize:   3,
			want:       idSet(fullAssigned, fullUnassigned),
		},
		{
			// Offset paging over a listing whose order changes mid-read skips
			// a job each time an already delivered job moves to the end.
			name:      "Jobs edited while an incremental read pages are not skipped",
			assigned:  editedAssigned,
			since:     date(2026, 10, 1),
			until:     date(2026, 10, 9),
			editEvery: 3,
			editTo:    date(2026, 10, 8).Add(12 * time.Hour),
			want:      idSet(editedAssigned),
		},
		{
			name:     "More than 100000 jobs on one day stops with a clear error",
			assigned: makeFakeJobs("a", 101000, busyHour, busyHour.Add(time.Hour)),
			since:    date(2026, 8, 9),
			until:    date(2026, 8, 12),
			wantErr:  errJobsWindowStalled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			location := tt.location
			if location == nil {
				location = time.UTC
			}

			api := newFakeJobsAPI(tt.assigned, tt.unassigned, location)
			api.editEvery, api.editTo = tt.editEvery, tt.editTo
			server := api.server(t)

			params := common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id", "modifiedDate"),
				Since:      tt.since,
				Until:      tt.until,
				PageSize:   tt.pageSize,
			}
			if tt.cursor != "" {
				params.NextPage = common.NextPageToken(server.URL + tt.cursor)
			}

			outcome := readAll(t, server.URL, params)

			// No request may ever reach the offset ceiling.
			assert.Equal(t, api.outOfRangeHits, 0)

			if tt.wantErr != nil {
				assert.Assert(t, errors.Is(outcome.err, tt.wantErr), "got %v", outcome.err)

				return
			}

			assert.NilError(t, outcome.err)
			assert.Assert(t, outcome.done)
			assertAllSeen(t, outcome, tt.want)

			if tt.noDupes {
				assert.Equal(t, outcome.duplicates, 0)
			}

			t.Logf("jobs=%d rows=%d duplicates=%d requests=%d",
				len(tt.want), outcome.rows, outcome.duplicates, api.requests)
		})
	}
}
