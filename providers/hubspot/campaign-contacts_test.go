package hubspot

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestLookupMarketingCampaignContactsBoundsConcurrency guards ENG-4431: reading contacts
// for a page of campaigns sent all of its report calls at once (300 for 100 campaigns),
// which exceeded HubSpot's 110 calls per 10 seconds and failed the whole page with 429s.
func TestLookupMarketingCampaignContactsBoundsConcurrency(t *testing.T) {
	t.Parallel()

	const campaigns = 20

	var (
		inFlight    atomic.Int64
		maxInFlight atomic.Int64
		reportCalls atomic.Int64
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/reports/contacts/") {
			current := inFlight.Add(1)
			defer inFlight.Add(-1)

			for {
				seen := maxInFlight.Load()
				if current <= seen || maxInFlight.CompareAndSwap(seen, current) {
					break
				}
			}

			reportCalls.Add(1)
			time.Sleep(10 * time.Millisecond)
		}

		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()

	connector, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	identifiers := make([]string, campaigns)
	for index := range identifiers {
		identifiers[index] = fmt.Sprintf("00000000-0000-0000-0000-%012d", index)
	}

	if _, err := connector.lookupMarketingCampaignContacts(t.Context(), identifiers); err != nil {
		t.Fatalf("lookup failed: %v", err)
	}

	if got, want := reportCalls.Load(), int64(3*campaigns); got != want {
		t.Errorf("report calls = %d, want %d", got, want)
	}

	if got := maxInFlight.Load(); got > maxConcurrentCampaignContactFetches {
		t.Errorf("max report calls in flight = %d, want at most %d", got, maxConcurrentCampaignContactFetches)
	}
}
