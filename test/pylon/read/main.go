package main

import (
	"context"
	"fmt"
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

	verifyIncrementalRead(ctx, conn)

	slog.Info("Read operation completed successfully.")
}

func verifyIncrementalRead(ctx context.Context, conn *pylon.Connector) {
	// A narrow window used only to locate the freshly created issue quickly.
	recentWindow := time.Now().UTC().Add(-2 * time.Minute)
	title := "Ampersand incremental-read test " + time.Now().UTC().Format("20060102T150405Z")

	// 1. Create an issue.
	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "issues",
		RecordData: map[string]any{
			"title":           title,
			"body_html":       "<p>Created by the Ampersand incremental-read test.</p>",
			"requester_email": "ampersand-incremental-test@example.com",
		},
	})
	if err != nil {
		utils.Fail("could not create test issue", "error", err)
	}

	issueID, _ := createRes.Data["id"].(string)
	slog.Info("step 1: created test issue", "id", issueID, "title", title)

	// 2. Read it back to confirm the freshly created issue is returned
	created, found := findIssue(ctx, conn, issueID, recentWindow)
	if !found {
		deleteIssue(ctx, conn, issueID)
		utils.Fail("step 2: created issue was not returned by the first read", "id", issueID)
	}

	createdAt := mustParseTime(created["created_at"])
	slog.Info("step 2: first read OK, created issue is present", "id", issueID, "created_at", createdAt)

	// 3. Update the issue by closing it. The row changes but created_at does not.
	time.Sleep(3 * time.Second)

	if _, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "issues",
		RecordId:   issueID,
		RecordData: map[string]any{"state": "closed"},
	}); err != nil {
		deleteIssue(ctx, conn, issueID)
		utils.Fail("step 3: could not close test issue", "id", issueID, "error", err)
	}

	slog.Info("step 3: closed test issue", "id", issueID)

	// 4. Read it back again to capture the new updated_at.
	updated, found := findIssue(ctx, conn, issueID, recentWindow)
	if !found {
		deleteIssue(ctx, conn, issueID)
		utils.Fail("step 4: updated issue was not returned when reading the recent window", "id", issueID)
	}

	updatedAt := mustParseTime(updated["updated_at"])
	if !updatedAt.After(createdAt) {
		deleteIssue(ctx, conn, issueID)
		utils.Fail("step 4: updated_at did not advance past created_at; cannot demonstrate the fix",
			"created_at", createdAt, "updated_at", updatedAt)
	}

	// 5. Re-read with a watermark strictly between created_at and updated_at.
	watermark := createdAt.Add(updatedAt.Sub(createdAt) / 2)
	_, foundAfterUpdate := findIssue(ctx, conn, issueID, watermark)

	// Delete the test issue before asserting.
	deleteIssue(ctx, conn, issueID)

	if !foundAfterUpdate {
		utils.Fail("step 5: INCREMENTAL READ REGRESSION: issue updated after the watermark was not re-read",
			"id", issueID, "watermark", watermark, "created_at", createdAt, "updated_at", updatedAt)
	}

	slog.Info("step 5: incremental read verified, issue updated after the watermark was re-read",
		"id", issueID, "watermark", watermark, "created_at", createdAt, "updated_at", updatedAt)
}

func deleteIssue(ctx context.Context, conn *pylon.Connector, id string) {
	url := fmt.Sprintf("%s/issues/%s", conn.ProviderInfo().BaseURL, id)

	if _, err := conn.JSONHTTPClient().Delete(ctx, url); err != nil {
		slog.Warn("could not delete test issue", "id", id, "error", err)

		return
	}

	slog.Info("cleaned up test issue", "id", id)
}

// findIssue reads issues page by page, filtered to updates since the given time, and returns the
// raw record whose id matches, if any.
func findIssue(ctx context.Context, conn *pylon.Connector, id string, since time.Time) (map[string]any, bool) {
	var nextPage common.NextPageToken

	for {
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "issues",
			Fields:     connectors.Fields("id", "created_at", "updated_at", "state"),
			Since:      since,
			NextPage:   nextPage,
		})
		if err != nil {
			utils.Fail("error reading issues", "error", err)
		}

		for _, row := range res.Data {
			if row.Raw["id"] == id {
				return row.Raw, true
			}
		}

		if res.Done || res.NextPage == "" {
			return nil, false
		}

		nextPage = res.NextPage
	}
}

func mustParseTime(value any) time.Time {
	str, _ := value.(string)

	parsed, err := time.Parse(time.RFC3339, str)
	if err != nil {
		utils.Fail("could not parse timestamp", "value", value, "error", err)
	}

	return parsed.UTC()
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
