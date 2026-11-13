// Live check of the AccuLynx 429 retry (ENG-4388).
//
// Phase 1 bursts raw GETs with the plain API-key client (no retry) until AccuLynx
// answers 429, and prints the Retry-After header it sends.
// Phase 2 bursts connector reads, which go through the retry wrapper; every read
// must succeed, with "AccuLynx rate limited a request, retrying after delay" logged
// for the ones that hit a 429.
package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	probeURL = "https://api.acculynx.com/api/v2/jobs?pageSize=1"
	// Documented limit is 10/s per key and 30/s per IP; the flood runs well over both.
	floodWorkers = 40
	floodFor     = 90 * time.Second
	burstSize    = 50
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	reader := utils.MustCreateProvCredJSON(credscanning.LoadPath(providers.AccuLynx), false)
	raw := utils.NewAPIKeyClient(ctx, reader, providers.AccuLynx)

	slog.Info("Phase 1: raw flood without retry", "workers", floodWorkers, "max", floodFor)

	// Flood until AccuLynx pushes back, so the quota is spent going into phase 2.
	if !rawBurst(ctx, raw) {
		utils.Fail("no 429 from AccuLynx during the flood; cannot verify the retry")
	}

	// Phase 1 just spent the quota, so phase 2 starts rate limited.
	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("Phase 2: connector burst with retry", "reads", burstSize)

	start := time.Now()

	var (
		wg     sync.WaitGroup
		failed atomic.Int32
	)

	for range burstSize {
		wg.Go(func() {
			_, err := conn.Read(ctx, common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")})
			if err != nil {
				failed.Add(1)
				slog.Error("connector read failed", "error", err)
			}
		})
	}

	wg.Wait()

	slog.Info("Phase 2 done", "reads", burstSize, "failed", failed.Load(), "took", time.Since(start).Round(time.Second))

	if failed.Load() > 0 {
		os.Exit(1)
	}

	slog.Info("PASS: every connector read succeeded under rate limiting")
}

// rawBurst keeps floodWorkers goroutines sending GETs back to back for up to
// floodFor, stopping early at the first 429. It prints every header on the first
// 200 and the first 429 so we can see what AccuLynx sends, and reports whether a
// 429 came back.
func rawBurst(ctx context.Context, client common.AuthenticatedHTTPClient) bool {
	ctx, stop := context.WithTimeout(ctx, floodFor)
	defer stop()

	var (
		wg              sync.WaitGroup
		mu              sync.Mutex
		counts          = map[int]int{}
		seen200, hit429 bool
	)

	started := time.Now()

	for range floodWorkers {
		wg.Go(func() {
			for ctx.Err() == nil {
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)

				resp, err := client.Do(req)
				if err != nil {
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()

				mu.Lock()
				counts[resp.StatusCode]++

				switch {
				case resp.StatusCode == http.StatusOK && !seen200:
					seen200 = true
					slog.Info("Headers on a normal 200", "headers", resp.Header)
				case resp.StatusCode == http.StatusTooManyRequests && !hit429:
					hit429 = true
					slog.Info("Got 429", "after", time.Since(started).Round(time.Millisecond),
						"headers", resp.Header, "body", string(body))
					stop()
				}
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	slog.Info("Phase 1 status codes", "counts", counts, "took", time.Since(started).Round(time.Second))

	return hit429
}
