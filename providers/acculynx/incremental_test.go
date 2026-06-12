package acculynx

import (
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// TestPairedDateWindow verifies the defaulting logic that protects /jobs and
// /jobs/{id}/history against AccuLynx's HTTP 400 when only one of
// startDate / endDate is sent.
func TestPairedDateWindow(t *testing.T) { //nolint:funlen
	t.Parallel()

	fixedSince := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	fixedUntil := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	epoch := time.Unix(0, 0).UTC()

	tests := []struct {
		name                    string
		since, until            time.Time
		wantSince               time.Time
		wantUntilDefaultedToNow bool
	}{
		{
			name:      "both set — pass through unchanged",
			since:     fixedSince,
			until:     fixedUntil,
			wantSince: fixedSince,
		},
		{
			name:                    "only Since set — Until defaults to now (platform backfill shape)",
			since:                   fixedSince,
			until:                   time.Time{},
			wantSince:               fixedSince,
			wantUntilDefaultedToNow: true,
		},
		{
			name:      "only Until set — Since defaults to epoch",
			since:     time.Time{},
			until:     fixedUntil,
			wantSince: epoch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := time.Now().UTC()
			gotSince, gotUntil := pairedDateWindow(tt.since, tt.until)
			after := time.Now().UTC()

			if !gotSince.Equal(tt.wantSince) {
				t.Errorf("since = %v, want %v", gotSince, tt.wantSince)
			}

			if gotUntil.IsZero() {
				t.Fatal("until should never be zero after pairedDateWindow")
			}

			if tt.wantUntilDefaultedToNow {
				if gotUntil.Before(before) || gotUntil.After(after) {
					t.Errorf("until = %v, want time.Now()-ish (between %v and %v)",
						gotUntil, before, after)
				}
			} else if !tt.until.IsZero() && !gotUntil.Equal(tt.until) {
				t.Errorf("until = %v, want %v (unchanged)", gotUntil, tt.until)
			}
		})
	}
}

// TestApplyJobsIncrementalFilter_OnlySinceSet is the regression test for the
// production HTTP 400 caused by sending startDate without endDate. The bug
// shipped because no unit or live test exercised the platform's backfill
// param shape (Since set, Until zero).
func TestApplyJobsIncrementalFilter_OnlySinceSet(t *testing.T) {
	t.Parallel()

	u, err := urlbuilder.New("https://api.acculynx.com", "/api/v2/jobs")
	if err != nil {
		t.Fatalf("urlbuilder: %v", err)
	}

	params := common.ReadParams{
		ObjectName: "jobs",
		Since:      time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
		// Until is intentionally zero — matches platform backfill shape.
	}

	applyJobsIncrementalFilter(u, params)

	assertQueryParam(t, u, "dateFilterType", "ModifiedDate")
	assertQueryParam(t, u, "startDate", "2026-05-26")
	assertQueryParamPresent(t, u, "endDate")
}

// TestApplyHistoryDateWindow_OnlySinceSet mirrors the jobs regression test for
// /jobs/{id}/history. Same code pattern as applyJobsIncrementalFilter, same
// defensive defaulting required.
func TestApplyHistoryDateWindow_OnlySinceSet(t *testing.T) {
	t.Parallel()

	u, err := urlbuilder.New("https://api.acculynx.com", "/api/v2/jobs/some-id/history")
	if err != nil {
		t.Fatalf("urlbuilder: %v", err)
	}

	params := common.ReadParams{
		ObjectName: "jobs/history",
		Since:      time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
	}

	applyHistoryDateWindow(u, params)

	assertQueryParam(t, u, "startDate", "2026-05-26")
	assertQueryParamPresent(t, u, "endDate")
}

func assertQueryParam(t *testing.T, u *urlbuilder.URL, key, want string) {
	t.Helper()

	got, ok := u.GetFirstQueryParam(key)
	if !ok {
		t.Fatalf("query param %q missing; URL = %s", key, u.String())
	}

	if got != want {
		t.Errorf("query param %q = %q, want %q", key, got, want)
	}
}

func assertQueryParamPresent(t *testing.T, u *urlbuilder.URL, key string) {
	t.Helper()

	got, ok := u.GetFirstQueryParam(key)
	if !ok {
		t.Fatalf("query param %q missing; URL = %s", key, u.String())
	}

	if got == "" {
		t.Errorf("query param %q is present but empty; URL = %s", key, u.String())
	}
}
