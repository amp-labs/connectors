package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/pylon"
	connTest "github.com/amp-labs/connectors/test/pylon"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("id", "channels", "name"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "email", "avatar_url"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)

	readIssueScenarios(ctx, conn)

	slog.Info("Read operation completed successfully.")
}

// readIssueScenarios exercises every filter shape buildIssueBody can emit against live
// Pylon. Row counts rising across the widening windows show that the updated_at filter
// is not capped at 30 days the way GET /issues caps start_time/end_time.
//
// Failures are logged rather than fatal, so one run reports on every shape.
func readIssueScenarios(ctx context.Context, conn *pylon.Connector) {
	now := time.Now().UTC()

	scenarios := []struct {
		name     string
		verifies string
		since    time.Time
		until    time.Time
	}{
		{
			name:     "no window",
			verifies: "filter omitted entirely; full backfill reads all issues",
		},
		{
			name:     "since only",
			verifies: "lone lower bound; Pylon accepts an 'and' wrapping a single subfilter",
			since:    now.AddDate(0, 0, -7),
		},
		{
			name:     "until only",
			verifies: "lone upper bound; same single-subfilter shape",
			until:    now,
		},
		{
			name:     "since and until, 7 days",
			verifies: "ordinary incremental read; two subfilters under 'and'",
			since:    now.AddDate(0, 0, -7),
			until:    now,
		},
		{
			name:     "since and until, 60 days",
			verifies: "window wider than the 30 day cap GET /issues enforces",
			since:    now.AddDate(0, 0, -60),
			until:    now,
		},
		{
			name:     "since and until, 120 days",
			verifies: "window far past the cap; confirms no silent truncation at 30 days",
			since:    now.AddDate(0, 0, -120),
			until:    now,
		},
	}

	for _, scenario := range scenarios {
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "issues",
			Fields:     connectors.Fields("id", "created_at", "updated_at", "state"),
			Since:      scenario.since,
			Until:      scenario.until,
		})
		if err != nil {
			slog.Error("issues read FAILED",
				"scenario", scenario.name,
				"verifies", scenario.verifies,
				"error", err,
			)

			continue
		}

		slog.Info("issues read OK",
			"scenario", scenario.name,
			"verifies", scenario.verifies,
			"rows", res.Rows,
			"done", res.Done,
		)
	}
}
