package acculynx

import (
	"testing"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

func TestApplyJobsOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		object   string
		wantSort bool
	}{
		// The /jobs list, which jobs/* objects also page through, is always sorted.
		{object: objectJobs, wantSort: true},
		// Other lists and the jobs/* leaf endpoints keep AccuLynx's default order.
		{object: objectContacts, wantSort: false},
		{object: objectEstimates, wantSort: false},
		{object: "jobs/history", wantSort: false},
	}

	for _, tt := range tests {
		t.Run(tt.object, func(t *testing.T) {
			t.Parallel()

			url, err := urlbuilder.New("https://api.acculynx.com", "/api/v2/x")
			if err != nil {
				t.Fatal(err)
			}

			applyJobsOrdering(url, tt.object)

			if got := hasJobsOrdering(url); got != tt.wantSort {
				t.Errorf("applyJobsOrdering(%q): sorted = %v, want %v (url %s)", tt.object, got, tt.wantSort, url)
			}

			_, hasSortBy := url.GetFirstQueryParam(sortByParam)
			if !tt.wantSort && hasSortBy {
				t.Errorf("applyJobsOrdering(%q) added sortBy: %s", tt.object, url)
			}
		})
	}
}
